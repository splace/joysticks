package joysticks

import (
	"time"
	"math"
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
	hatPanXEvents         map[uint8]chan event
	hatPanYEvents         map[uint8]chan event
	hatPositionEvents     map[uint8]chan event
	hatAngleEvents        map[uint8]chan event
}

type event interface {
	Moment() time.Duration
}

type when struct {
	time time.Duration
}

func (b when) Moment() time.Duration {
	return b.time
}

type HatPositionEvent struct {
	when
	X, Y float32
}

type ButtonChangeEvent struct {
	when
}

type HatPanXEvent struct {
	when
	V float32
}

type HatPanYEvent struct {
	when
	V float32
}

type HatAngleEvent struct {
	when
	Angle float32
}

// ParcelOutEvents interprets whats on the State.OSEvent channel, then puts the required event(s), on any registered channel(s).
func (js Joystick) ParcelOutEvents() {
	for {
		if evt, ok := <-js.OSEvent; ok {
			switch evt.Type {
			case 1:
				if evt.Value == 0 {
					if c, ok := js.buttonOpenEvents[js.buttons[evt.Index].number]; ok {
						c <- ButtonChangeEvent{when{toDuration(evt.Time)}}
					}
					if c, ok := js.buttonLongPressEvents[js.buttons[evt.Index].number]; ok {
						if toDuration(evt.Time) > js.buttons[evt.Index].time+time.Second {
							c <- ButtonChangeEvent{when{toDuration(evt.Time)}}
						}
					}
				}
				if evt.Value == 1 {
					if c, ok := js.buttonCloseEvents[js.buttons[evt.Index].number]; ok {
						c <- ButtonChangeEvent{when{toDuration(evt.Time)}}
					}
				}
				js.buttons[evt.Index] = button{js.buttons[evt.Index].number, toDuration(evt.Time), evt.Value != 0}
			case 2:
				switch js.hatAxes[evt.Index].axis {
				case 1:
					if c, ok := js.hatPanXEvents[js.hatAxes[evt.Index].number]; ok {
						c <- HatPanXEvent{when{toDuration(evt.Time)}, float32(evt.Value) / maxValue}
					}
				case 2:
					if c, ok := js.hatPanYEvents[js.hatAxes[evt.Index].number]; ok {
						c <- HatPanYEvent{when{toDuration(evt.Time)}, float32(evt.Value) / maxValue}
					}
				}
				if c, ok := js.hatPositionEvents[js.hatAxes[evt.Index].number]; ok {
					switch js.hatAxes[evt.Index].axis {
					case 1:
						c <- HatPositionEvent{when{toDuration(evt.Time)}, float32(evt.Value) / maxValue, js.hatAxes[evt.Index+1].value}
					case 2:
						c <- HatPositionEvent{when{toDuration(evt.Time)}, js.hatAxes[evt.Index-1].value, float32(evt.Value) / maxValue}
					}
				}
				if c, ok := js.hatAngleEvents[js.hatAxes[evt.Index].number]; ok {
					switch js.hatAxes[evt.Index].axis {
					case 1:
						c <- HatAngleEvent{when{toDuration(evt.Time)}, float32(math.Atan2(float64(evt.Value),float64(js.hatAxes[evt.Index+1].value)))}
					case 2:
						c <- HatAngleEvent{when{toDuration(evt.Time)}, float32(math.Atan2(float64(js.hatAxes[evt.Index-1].value), float64(evt.Value) / maxValue))}
					}
				}
				js.hatAxes[evt.Index] = hatAxis{js.hatAxes[evt.Index].number, js.hatAxes[evt.Index].axis, toDuration(evt.Time), float32(evt.Value) / maxValue}
			default:
				// log.Println("unknown input type. ",evt.Type & 0x7f)
			}
		} else {
			break
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
	js.hatPositionEvents[hat] = c
	return c
}

// hat axis-X moved
func (js Joystick) OnPanX(hat uint8) chan event {
	c := make(chan event)
	js.hatPanXEvents[hat] = c
	return c
}

// hat axis-Y moved
func (js Joystick) OnPanY(hat uint8) chan event {
	c := make(chan event)
	js.hatPanYEvents[hat] = c
	return c
}

// hat axis-Y moved
func (js Joystick) OnRotate(hat uint8) chan event {
	c := make(chan event)
	js.hatAngleEvents[hat] = c
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
