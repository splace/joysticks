package joysticks

import (
	"math"
	"time"
	//	"fmt"
)

// TODO drag event

var LongPressDelay = time.Second/2

type hatAxis struct {
	number   uint8
	axis     uint8
	reversed bool
	time     time.Duration
	value    float32
}

type button struct {
	number uint8
	time   time.Duration
	value  bool
}


type eventType uint8

const (
	buttonChange eventType = iota
	buttonClose
	buttonOpen
	buttonLongPress
	hatChange
	hatPanX
	hatPanY
	hatPosition
	hatAngle
	hatRadius
	hatCentered
	hatEdge
)

type eventSigniture struct {
	typ eventType
	number uint8
}


// HID holds the in-coming event channel, available button and hat indexes, and registered events, for a human interface device.
// has methods to control and adjust behaviour.
type HID struct {
	OSEvents              chan osEventRecord
	Buttons               map[uint8]button
	HatAxes               map[uint8]hatAxis
	Events	map[eventSigniture]chan Event
}

// Events always have the time they occurred.
type Event interface {
	Moment() time.Duration
}

type when struct {
	time time.Duration
}

func (b when) Moment() time.Duration {
	return b.time
}

// button changed
type ButtonEvent struct {
	when
	number uint8
	value  bool
}

// hat changed
type HatEvent struct {
	when
	number uint8
	axis   uint8
	value  float32
}

// Hat position event type. X,Y{-1...1}
type CoordsEvent struct {
	when
	X, Y float32
}

// Hat Axis event type. V{-1...1}
type PanEvent struct {
	when
	V float32
}

// Hat angle event type. Angle{-Pi...Pi}
type AngleEvent struct {
	when
	Angle float32
}

// Hat radius event type. Radius{0...√2}
type RadiusEvent struct {
	when
	Radius float32
}

// ParcelOutEvents waits on the HID.OSEvent channel (so is blocking), then puts the required event(s), on any registered channel(s).
func (d HID) ParcelOutEvents() {
	for {
		if evt, ok := <-d.OSEvents; ok {
			switch evt.Type {
			case 1:
				b := d.Buttons[evt.Index]
				if c, ok := d.Events[eventSigniture{buttonChange,b.number}]; ok {
					c <- ButtonEvent{when{toDuration(evt.Time)}, b.number, evt.Value == 1}
				}
				if evt.Value == 0 {
					if c, ok := d.Events[eventSigniture{buttonOpen,b.number}]; ok {
						c <- when{toDuration(evt.Time)}
					}
					if c, ok := d.Events[eventSigniture{buttonLongPress,b.number}]; ok {
						if toDuration(evt.Time) > b.time+LongPressDelay {
							c <- when{toDuration(evt.Time)}
						}
					}
				}
				if evt.Value == 1 {
					if c, ok := d.Events[eventSigniture{buttonClose,b.number}]; ok {
						c <- when{toDuration(evt.Time)}
					}
				}
				d.Buttons[evt.Index] = button{b.number, toDuration(evt.Time), evt.Value != 0}
			case 2:
				h := d.HatAxes[evt.Index]
				v := float32(evt.Value) / maxValue
				if h.reversed {
					v = -v
				}
				if c, ok := d.Events[eventSigniture{hatChange,h.number}]; ok {
					c <- HatEvent{when{toDuration(evt.Time)}, h.number, h.axis, v}
				}
				switch h.axis {
				case 1:
					if c, ok := d.Events[eventSigniture{hatPanX,h.number}]; ok {
						c <- PanEvent{when{toDuration(evt.Time)}, v}
					}
				case 2:
					if c, ok := d.Events[eventSigniture{hatPanY,h.number}]; ok {
						c <- PanEvent{when{toDuration(evt.Time)}, v}
					}
				}
				if c, ok := d.Events[eventSigniture{hatPosition,h.number}]; ok {
					switch h.axis {
					case 1:
						c <- CoordsEvent{when{toDuration(evt.Time)}, v, d.HatAxes[evt.Index+1].value}
					case 2:
						c <- CoordsEvent{when{toDuration(evt.Time)}, d.HatAxes[evt.Index-1].value, v}
					}
				}
				if c, ok := d.Events[eventSigniture{hatAngle,h.number}]; ok {
					switch h.axis {
					case 1:
						c <- AngleEvent{when{toDuration(evt.Time)}, float32(math.Atan2(float64(v), float64(d.HatAxes[evt.Index+1].value)))}
					case 2:
						c <- AngleEvent{when{toDuration(evt.Time)}, float32(math.Atan2(float64(d.HatAxes[evt.Index-1].value), float64(v)))}
					}
				}
				if c, ok := d.Events[eventSigniture{hatRadius,h.number}]; ok {
					switch h.axis {
					case 1:
						c <- RadiusEvent{when{toDuration(evt.Time)}, float32(math.Sqrt(float64(v)*float64(v) + float64(d.HatAxes[evt.Index+1].value)*float64(d.HatAxes[evt.Index+1].value)))}
					case 2:
						c <- RadiusEvent{when{toDuration(evt.Time)}, float32(math.Sqrt(float64(d.HatAxes[evt.Index-1].value)*float64(d.HatAxes[evt.Index-1].value) + float64(v)*float64(v)))}
					}
				}
				if c, ok := d.Events[eventSigniture{hatEdge,h.number}]; ok {
					// fmt.Println(v,h)
					if (v == 1 || v == -1) && h.value != 1 && h.value != -1 {
						switch h.axis {
						case 1:
							c <- AngleEvent{when{toDuration(evt.Time)}, float32(math.Atan2(float64(v), float64(d.HatAxes[evt.Index+1].value)))}
						case 2:
							c <- AngleEvent{when{toDuration(evt.Time)}, float32(math.Atan2(float64(d.HatAxes[evt.Index-1].value), float64(v)))}
						}
					}
				}
				if c, ok := d.Events[eventSigniture{hatCentered,h.number}]; ok {
					if v == 0 && h.value != 0 {
						switch h.axis {
						case 1:
							if d.HatAxes[evt.Index+1].value == 0 {
								c <- when{toDuration(evt.Time)}
							}
						case 2:
							if d.HatAxes[evt.Index-1].value == 0 {
								c <- when{toDuration(evt.Time)}
							}
						}
					}
				}
				d.HatAxes[evt.Index] = hatAxis{h.number, h.axis, h.reversed, toDuration(evt.Time), v}
			default:
				// log.Println("unknown input type. ",evt.Type & 0x7f)
			}
		} else {
			break
		}
	}
}

// Type of register-able methods and the index they are called with. (Note: the event type is indicated by the method.)
type Channel struct {
	Number uint8
	Method func(HID, uint8) chan Event
}


// Capture is highlevel automation of the setup of event channels.
// returned are matching chan's for each registree, which then receive events of the type and index the registree indicated.
// It uses the first available joystick, from a max of 4.
// Doesn't return HID object, so settings are fixed.
func Capture(registrees ...Channel) []chan Event {
	d := Connect(1)
	for i := 2; d == nil && i < 5; i++ {
		d = Connect(i)
	}
	if d == nil {
		return nil
	}
	go d.ParcelOutEvents()
	chans := make([]chan Event, len(registrees))
	for i, fns := range registrees {
		chans[i] = fns.Method(*d, fns.Number)
	}
	return chans
}


// button changes event channel.
func (d HID) OnButton(button uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSigniture{buttonChange,button}] = c
	return c
}

// button goes open event channel.
func (d HID) OnOpen(button uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSigniture{buttonOpen,button}] = c
	return c
}

// button goes closed event channel.
func (d HID) OnClose(button uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSigniture{buttonClose,button}] = c
	return c
}

// button goes open and the previous event, closed, was more than LongPressDelay ago, event channel.
func (d HID) OnLong(button uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSigniture{buttonLongPress,button}] = c
	return c
}

// hat moved event channel.
func (d HID) OnHat(hat uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSigniture{hatChange,hat}] = c
	return c
}

// hat position changed event channel.
func (d HID) OnMove(hat uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSigniture{hatPosition,hat}] = c
	return c
}

// hat axis-X moved event channel.
func (d HID) OnPanX(hat uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSigniture{hatPanX,hat}] = c
	return c
}

// hat axis-Y moved event channel.
func (d HID) OnPanY(hat uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSigniture{hatPanY,hat}] = c
	return c
}

// hat angle changed event channel.
func (d HID) OnRotate(hat uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSigniture{hatAngle,hat}] = c
	return c
}

// hat moved event channel.
func (d HID) OnCenter(hat uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSigniture{hatCentered,hat}] = c
	return c
}

// hat moved to edge
func (d HID) OnEdge(hat uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSigniture{hatEdge,hat}] = c
	return c
}

// see if Button exists.
func (d HID) ButtonExists(button uint8) (ok bool) {
	for _, v := range d.Buttons {
		if v.number == button {
			return true
		}
	}
	return
}

// see if Hat exists.
func (d HID) HatExists(hat uint8) (ok bool) {
	for _, v := range d.HatAxes {
		if v.number == hat {
			return true
		}
	}
	return
}

// Button current state.
func (d HID) ButtonClosed(button uint8) bool {
	return d.Buttons[button].value
}

// Hat latest position. (coords slice needs to be long enough to hold all axis.)
func (d HID) HatCoords(hat uint8, coords []float32) {
	for _, h := range d.HatAxes {
		if h.number == hat {
			coords[h.axis-1] = h.value
		}
	}
	return
}

// insert events as if from hardware.
func (d HID) InsertSyntheticEvent(v int16, t uint8, i uint8) {
	d.OSEvents <- osEventRecord{Value: v, Type: t, Index: i}
}

