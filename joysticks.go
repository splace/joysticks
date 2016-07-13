package joysticks

import (
	"time"
)

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

//Joystick holds the in-coming event channel, mappings, and registered events for a joystick, and has methods to control and adjust behaviour.
type Joystick struct {
	OSEvent               chan osEventRecord
	buttons               map[uint8]button
	hatAxes               map[uint8]hatAxis
	buttonCloseEvents     map[uint8]chan event
	buttonOpenEvents      map[uint8]chan event
	buttonLongPressEvents map[uint8]chan event
	hatChangeEvents       map[uint8]chan event
}

type event interface {
	Moment() time.Duration
}

type HatChangeEvent struct {
	time time.Duration
	X, Y float32
}

func (b HatChangeEvent) Moment() time.Duration {
	return b.time
}

type ButtonChangeEvent struct {
	time time.Duration
}

func (b ButtonChangeEvent) Moment() time.Duration {
	return b.time
}

// ParcelOutEvents interprets whats on the State.OSEvent channel, then puts the required event(s), on any registered channel(s).
func (js Joystick) ParcelOutEvents() {
	for {
		evt, ok := <-js.OSEvent
		if !ok {
			break
		}
		switch evt.Type {
		case 1:
			if evt.Value == 0 {
				if c, ok := js.buttonOpenEvents[js.buttons[evt.Index].number]; ok {
					c <- ButtonChangeEvent{toDuration(evt.Time)}
				}
				if c, ok := js.buttonLongPressEvents[js.buttons[evt.Index].number]; ok {
					if toDuration(evt.Time) > js.buttons[evt.Index].time+time.Second {
						c <- ButtonChangeEvent{toDuration(evt.Time)}
					}
				}
			}
			if evt.Value == 1 {
				if c, ok := js.buttonCloseEvents[js.buttons[evt.Index].number]; ok {
					c <- ButtonChangeEvent{toDuration(evt.Time)}
				}
			}
			js.buttons[evt.Index] = button{js.buttons[evt.Index].number, toDuration(evt.Time), evt.Value != 0}
		case 2:
			if c, ok := js.hatChangeEvents[js.hatAxes[evt.Index].number]; ok {
				switch js.hatAxes[evt.Index].axis {
				case 1:
					c <- HatChangeEvent{toDuration(evt.Time), float32(evt.Value) / maxValue, js.hatAxes[evt.Index+1].value}
				case 2:
					c <- HatChangeEvent{toDuration(evt.Time), js.hatAxes[evt.Index-1].value, float32(evt.Value) / maxValue}
				}
			}
			js.hatAxes[evt.Index] = hatAxis{js.hatAxes[evt.Index].number, js.hatAxes[evt.Index].axis, toDuration(evt.Time), float32(evt.Value) / maxValue}
		default:
			// log.Println("unknown input type. ",evt.Type & 0x7f)
		}
	}
}

// Type of registerable methods and the index they are called with. (Note: the event type is indicated by the method.)
type Channel struct {
	Number uint8
	Method func(Joystick, uint8) chan event
}

// button goes open
func (js Joystick) OnOpen(button uint8) chan event {
	c := make(chan event)
	js.buttonOpenEvents[button] = c
	return c
}

// button goes closed
func (js Joystick) OnClose(button uint8) chan event {
	c := make(chan event)
	js.buttonCloseEvents[button] = c
	return c
}

// button goes open and last event on it, closed, wasn't recent. (within 1 second)
func (js Joystick) OnLong(button uint8) chan event {
	c := make(chan event)
	js.buttonLongPressEvents[button] = c
	return c
}

// hat moved
func (js Joystick) OnMove(hat uint8) chan event {
	c := make(chan event)
	js.hatChangeEvents[hat] = c
	return c
}

func (js Joystick) ButtonExists(button uint8) (ok bool) {
	for _, v := range js.buttons {
		if v.number == button {
			return true
		}
	}
	return
}

func (js Joystick) HatExists(hat uint8) (ok bool) {
	for _, v := range js.hatAxes {
		if v.number == hat {
			return true
		}
	}
	return
}

func (js Joystick) InsertSyntheticEvent(v int16, t uint8, i uint8) {
	js.OSEvent <- osEventRecord{Value: v, Type: t, Index: i}
}
