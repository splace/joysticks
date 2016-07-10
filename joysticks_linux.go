// +build linux

package joysticks

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

// see; https://www.kernel.org/doc/Documentation/input/joystick-api.txt
type osEventRecord struct {
	Time  uint32 // event timestamp in milliseconds 32bit so, about a month
	Value int16  // value
	Type  uint8  // event type
	Index uint8  // axis/button
}

const maxValue = 1<<15 - 1

type hatAxis struct {
	number uint8
	axis   uint8
	time   time.Duration
	value  float32
}

type button struct {
	number uint8
	time   time.Duration
	value  bool
}

type state struct {
	osEvent           chan osEventRecord
	buttons           map[uint8]button
	hatAxes           map[uint8]hatAxis
	buttonCloseEvents map[uint8]chan event
	buttonOpenEvents  map[uint8]chan event
	hatChangeEvents   map[uint8]chan event
}

type event interface {
	Moment() time.Duration
}

type hatChangeEvent struct {
	time time.Duration
	x, y float32
}

func (b hatChangeEvent) Moment() time.Duration {
	return b.time
}

type buttonChangeEvent struct {
	time time.Duration
}

func (b buttonChangeEvent) Moment() time.Duration {
	return b.time
}

type channel struct {
	number    uint8
	channelFn func(state, uint8) chan event
}

// Capture arranges for the registrees (channels) to get particular events.
// Intended for bacic use since doesn't return state object.
func Capture(registrees ...channel) []chan event {
	js, err := Connect(1)
	if err != nil {
		return nil
	}
	go js.ProcessEvents()
	chans := make([]chan event, len(registrees))
	for i, fns := range registrees {
		chans[i] = fns.channelFn(js, fns.number)
	}
	return chans
}

// Connect sets up a go routine that puts a joysticks events onto registered channels.
// to register channels use the returned state object's On<xxx>(index) methods.
func Connect(index int) (js state, e error) {
	r, e := os.OpenFile(fmt.Sprintf("/dev/input/js%d", index-1), os.O_RDONLY, 0666)
	if e != nil {
		return
	}
	js = state{make(chan osEventRecord), make(map[uint8]button), make(map[uint8]hatAxis), make(map[uint8]chan event), make(map[uint8]chan event), make(map[uint8]chan event)}
	// start thread to read joystick eventd to the joystick.state osEvent channel
	go eventPipe(r, js.osEvent)
	js.populate()
	return js, nil
}

// fill in the joysticks available events from the synthetic state events burst produced initially by the driver.
func (js state) populate() {
	for buttonNumber, hatNumber, axisNumber := 1, 1, 1; ; {
		evt := <-js.osEvent
		switch evt.Type {
		case 0x81:
			js.buttons[evt.Index] = button{uint8(buttonNumber), toDuration(evt.Time), evt.Value != 0}
			buttonNumber += 1
		case 0x82:
			js.hatAxes[evt.Index] = hatAxis{uint8(hatNumber), uint8(axisNumber), toDuration(evt.Time), float32(evt.Value) / maxValue}
			axisNumber += 1
			if axisNumber > 2 {
				axisNumber = 1
				hatNumber += 1
			}
		default:
			go func() { js.osEvent <- evt }() // put the consumed, first, after end of synthetic burst, real event, back on channel.
			return
		}
	}
	return
}

// pipe any readable events onto channel.
func eventPipe(r io.Reader, c chan osEventRecord) {
	var evt osEventRecord
	for {
		err := binary.Read(r, binary.LittleEndian, &evt)
		if err != nil {
			close(c)
			return
		}
		c <- evt
	}
}

// interpret whats appearing on osEvent channel, then put, on any required out channel(s), the requisite event.
func (js state) ProcessEvents() {
	for {
		evt, ok := <-js.osEvent
		if !ok {
			break
		}
		switch evt.Type {
		case 1:
			if evt.Value == 0 {
				if c, ok := js.buttonOpenEvents[js.buttons[evt.Index].number]; ok {
					c <- buttonChangeEvent{toDuration(evt.Time)}
				}
			}
			if evt.Value == 1 {
				if c, ok := js.buttonCloseEvents[js.buttons[evt.Index].number]; ok {
					c <- buttonChangeEvent{toDuration(evt.Time)}
				}
			}
			js.buttons[evt.Index] = button{js.buttons[evt.Index].number, toDuration(evt.Time), evt.Value != 0}
		case 2:
			if c, ok := js.hatChangeEvents[js.hatAxes[evt.Index].number]; ok {
				switch js.hatAxes[evt.Index].axis {
				case 1:
					c <- hatChangeEvent{toDuration(evt.Time), float32(evt.Value) / maxValue, js.hatAxes[evt.Index+1].value}
				case 2:
					c <- hatChangeEvent{toDuration(evt.Time), js.hatAxes[evt.Index-1].value, float32(evt.Value) / maxValue}
				}
			}
			js.hatAxes[evt.Index] = hatAxis{js.hatAxes[evt.Index].number, js.hatAxes[evt.Index].axis, toDuration(evt.Time), float32(evt.Value) / maxValue}
		default:
			// log.Println("unknown input type. ",evt.Type & 0x7f)
		}
	}
}

func toDuration(m uint32) time.Duration {
	return time.Duration(m) * 1000000
}

func (js state) OnOpen(button uint8) (c chan event) {
	c = make(chan event)
	js.buttonOpenEvents[button] = c
	return c
}

func (js state) OnClose(button uint8) (c chan event) {
	c = make(chan event)
	js.buttonCloseEvents[button] = c
	return c
}

func (js state) OnMove(hatAsix uint8) (c chan event) {
	c = make(chan event)
	js.hatChangeEvents[hatAsix] = c
	return c
}

func (js state) ButtonExists(button uint8) (ok bool) {
	for _, v := range js.buttons {
		if v.number == button {
			return true
		}
	}
	return
}

func (js state) HatExists(hatAxis uint8) (ok bool) {
	for _, v := range js.hatAxes {
		if v.number == hatAxis {
			return true
		}
	}
	return
}

func (js state) InsertSyntheticEvent(v int16, t uint8, i uint8) {
	js.osEvent <- osEventRecord{Value: v, Type: t, Index: i}
}


