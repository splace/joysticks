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
	Time  uint32 // event timestamp, unknown base, in milliseconds 32bit so, about a month
	Value int16  // value
	Type  uint8  // event type
	Index uint8  // axis/button
}

const maxValue = 1<<15 - 1

// Capture returns a chan, for each registree, getting the events the registree indicates.
// Finds the first unused joystick, from a max of 4.
// Intended for bacic use since doesn't return state object.
func Capture(registrees ...Channel) []chan event {
	js := Connect(1)
	for i := 2; js == nil && i < 5; i++ {
		js = Connect(i)
	}
	if js == nil {
		return nil
	}
	go js.ParcelOutEvents()
	chans := make([]chan event, len(registrees))
	for i, fns := range registrees {
		chans[i] = fns.Method(*js, fns.Number)
	}
	return chans
}

// Connect sets up a go routine that puts a joysticks events onto registered channels.
// register channels by using the returned state object's On<xxx>(index) methods.
// Note: only one event, of each type '<xxx>', for each 'index', re-registering stops events going on the old channel.
// then activate using state objects ParcelOutEvents() method.(usually in a go routine.)
func Connect(index int) (js *State) {
	r, e := os.OpenFile(fmt.Sprintf("/dev/input/js%d", index-1), os.O_RDWR, 0400)
	if e != nil {
		return nil
	}
	js = &State{make(chan osEventRecord), make(map[uint8]button), make(map[uint8]hatAxis), make(map[uint8]chan event), make(map[uint8]chan event), make(map[uint8]chan event), make(map[uint8]chan event)}
	// start thread to read joystick events to the joystick.state osEvent channel
	go eventPipe(r, js.OSEvent)
	js.populate()
	return js
}

// fill in the joysticks available events from the synthetic state events burst produced initially by the driver.
func (js State) populate() {
	for buttonNumber, hatNumber, axisNumber := 1, 1, 1; ; {
		evt := <-js.OSEvent
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
			go func() { js.OSEvent <- evt }() // put the consumed, first, after end of synthetic burst, real event, back on channel.
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
func toDuration(m uint32) time.Duration {
	return time.Duration(m) * 1000000
}


