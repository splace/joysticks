package joysticks

import (
	"math"
	"time"
	//"fmt"
)

var LongPressDelay = time.Second / 2
var DoublePressDelay = time.Second / 10

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
	buttonDoublePress
	hatChange
	hatPanX
	hatPanY
	hatPosition
	hatAngle
	hatRadius
	hatCentered
	hatEdge
	hatVelocityX
	hatVelocityY
)

// signature of an event
type eventSignature struct {
	eventType
	number uint8
}

// HID holds the in-coming event channel, available button and hat indexes, and registered events, for a human interface device.
// It has methods to control and adjust behaviour.
type HID struct {
	OSEvents chan osEventRecord
	Buttons  map[uint8]button
	HatAxes  map[uint8]hatAxis
	Events   map[eventSignature]chan Event
}

// Events always have the time they occurred.
type Event interface {
	Moment() time.Duration
}

type when struct {
	Time time.Duration
}

func (b when) Moment() time.Duration {
	return b.Time
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
type AxisEvent struct {
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

// ParcelOutEvents waits on the HID.OSEvents channel (so is blocking), then puts any events matching onto any registered channel(s).
func (d HID) ParcelOutEvents() {
	for evt := range d.OSEvents {
		switch evt.Type {
		case 1:
			b := d.Buttons[evt.Index]
			if c, ok := d.Events[eventSignature{buttonChange, b.number}]; ok {
				c <- ButtonEvent{when{toDuration(evt.Time)}, b.number, evt.Value == 1}
			}
			if evt.Value == 0 {
				if c, ok := d.Events[eventSignature{buttonOpen, b.number}]; ok {
					c <- when{toDuration(evt.Time)}
				}
				if c, ok := d.Events[eventSignature{buttonLongPress, b.number}]; ok {
					if toDuration(evt.Time) > b.time+LongPressDelay {
						c <- when{toDuration(evt.Time)}
					}
				}
			}
			if evt.Value == 1 {
				if c, ok := d.Events[eventSignature{buttonClose, b.number}]; ok {
					c <- when{toDuration(evt.Time)}
				}
				if c, ok := d.Events[eventSignature{buttonDoublePress, b.number}]; ok {
					if toDuration(evt.Time) < b.time+DoublePressDelay {
						c <- when{toDuration(evt.Time)}
					}
				}
			}
			d.Buttons[evt.Index] = button{b.number, toDuration(evt.Time), evt.Value != 0}
		case 2:
			h := d.HatAxes[evt.Index]
			v := float32(evt.Value) / maxValue
			if h.reversed {
				v = -v
			}
			if c, ok := d.Events[eventSignature{hatChange, h.number}]; ok {
				c <- HatEvent{when{toDuration(evt.Time)}, h.number, h.axis, v}
			}
			switch h.axis {
			case 1:
				if c, ok := d.Events[eventSignature{hatPanY, h.number}]; ok {
					c <- AxisEvent{when{toDuration(evt.Time)}, v}
				}
				if c, ok := d.Events[eventSignature{hatVelocityY, h.number}]; ok {
					c <- AxisEvent{when{toDuration(evt.Time)}, (v-d.HatAxes[evt.Index].value)/float32((toDuration(evt.Time)-d.HatAxes[evt.Index].time).Seconds())}
				}
			case 2:
				if c, ok := d.Events[eventSignature{hatPanX, h.number}]; ok {
					c <- AxisEvent{when{toDuration(evt.Time)}, v}
				}
				if c, ok := d.Events[eventSignature{hatVelocityX, h.number}]; ok {
					c <- AxisEvent{when{toDuration(evt.Time)}, (v-d.HatAxes[evt.Index].value)/float32((toDuration(evt.Time)-d.HatAxes[evt.Index].time).Seconds())}
				}
			}
			if c, ok := d.Events[eventSignature{hatPosition, h.number}]; ok {
				switch h.axis {
				case 1:
					c <- CoordsEvent{when{toDuration(evt.Time)},  v ,d.HatAxes[evt.Index+1].value}
				case 2:
					c <- CoordsEvent{when{toDuration(evt.Time)},  d.HatAxes[evt.Index-1].value,v}
				}
			}
			if c, ok := d.Events[eventSignature{hatAngle, h.number}]; ok {
				switch h.axis {
				case 1:
					c <- AngleEvent{when{toDuration(evt.Time)}, float32(math.Atan2(float64(d.HatAxes[evt.Index+1].value), float64(v)))}
				case 2:
					c <- AngleEvent{when{toDuration(evt.Time)}, float32(math.Atan2(float64(v), float64(d.HatAxes[evt.Index-1].value)))}
				}
			}
			if c, ok := d.Events[eventSignature{hatRadius, h.number}]; ok {
				switch h.axis {
				case 1:
					c <- RadiusEvent{when{toDuration(evt.Time)}, float32(math.Sqrt(float64(d.HatAxes[evt.Index+1].value)*float64(d.HatAxes[evt.Index+1].value) + float64(v)*float64(v)))}
				case 2:
					c <- RadiusEvent{when{toDuration(evt.Time)}, float32(math.Sqrt(float64(v)*float64(v) + float64(d.HatAxes[evt.Index-1].value)*float64(d.HatAxes[evt.Index-1].value)))}
				}
			}
			if c, ok := d.Events[eventSignature{hatEdge, h.number}]; ok {
				// fmt.Println(v,h)
				if (v == 1 || v == -1) && h.value != 1 && h.value != -1 {
					switch h.axis {
					case 1:
						c <- AngleEvent{when{toDuration(evt.Time)}, float32(math.Atan2(float64(d.HatAxes[evt.Index+1].value), float64(v)))}
					case 2:
						c <- AngleEvent{when{toDuration(evt.Time)}, float32(math.Atan2(float64(v), float64(d.HatAxes[evt.Index-1].value)))}
					}
				}
			}
			if c, ok := d.Events[eventSignature{hatCentered, h.number}]; ok {
				if v == 0 && h.value != 0 {
					switch h.axis {
					case 2:
						if d.HatAxes[evt.Index-1].value == 0 {
							c <- when{toDuration(evt.Time)}
						}
					case 1:
						if d.HatAxes[evt.Index+1].value == 0 {
							c <- when{toDuration(evt.Time)}
						}
					}
				}
			}
			d.HatAxes[evt.Index] = hatAxis{h.number, h.axis, h.reversed, toDuration(evt.Time), v}
		default:
			// log.Println("unknown input type. ",evt.Type & 0x7f)
		}
	}
}

// Type of register-able methods and the index they are called with. (Note: the event type is indicated by the method.)
type Channel struct {
	Number uint8
	Method func(HID, uint8) chan Event
}

// Capture is highlevel automation of the setup of event channels.
// Returned is a slice of chan's, matching each registree, which then receive events of the type and index the registree indicated.
// It uses the first available joystick, from a max of 4.
// Since it doesn't return a HID object, channels are immutable.
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
func (d HID) OnButton(index uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSignature{buttonChange, index}] = c
	return c
}

// button goes open event channel.
func (d HID) OnOpen(index uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSignature{buttonOpen, index}] = c
	return c
}

// button goes closed event channel.
func (d HID) OnClose(index uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSignature{buttonClose, index}] = c
	return c
}

// button goes open and the previous event, closed, was more than LongPressDelay ago, event channel.
func (d HID) OnLong(index uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSignature{buttonLongPress, index}] = c
	return c
}

// button goes closed and the previous event, open, was less than DoublePressDelay ago, event channel.
func (d HID) OnDouble(index uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSignature{buttonDoublePress, index}] = c
	return c
}

// hat moved event channel.
func (d HID) OnHat(index uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSignature{hatChange, index}] = c
	return c
}

// hat position changed event channel.
func (d HID) OnMove(index uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSignature{hatPosition, index}] = c
	return c
}

// hat axis-X moved event channel.
func (d HID) OnPanX(index uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSignature{hatPanX, index}] = c
	return c
}

// hat axis-Y moved event channel.
func (d HID) OnPanY(index uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSignature{hatPanY, index}] = c
	return c
}

// hat axis-X speed changed event channel.
func (d HID) OnSpeedX(index uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSignature{hatVelocityX, index}] = c
	return c
}

// hat axis-Y speed changed event channel.
func (d HID) OnSpeedY(index uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSignature{hatVelocityY, index}] = c
	return c
}

// hat angle changed event channel.
func (d HID) OnRotate(index uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSignature{hatAngle, index}] = c
	return c
}

// hat moved event channel.
func (d HID) OnCenter(index uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSignature{hatCentered, index}] = c
	return c
}

// hat moved to edge
func (d HID) OnEdge(index uint8) chan Event {
	c := make(chan Event)
	d.Events[eventSignature{hatEdge, index}] = c
	return c
}

// hat integrate
//func (d HID) OnIntegrate(c Channel) chan Event {
//	var e,le Event
//	e:=Event{}
//	c := make(chan Event)
//	d.Events[eventSignature{hatEdge, hat}] = c
//	return c
//}


// see if Button exists.
func (d HID) ButtonExists(index uint8) (ok bool) {
	for _, v := range d.Buttons {
		if v.number == index {
			return true
		}
	}
	return
}

// see if Hat exists.
func (d HID) HatExists(index uint8) (ok bool) {
	for _, v := range d.HatAxes {
		if v.number == index {
			return true
		}
	}
	return
}

// Button current state.
func (d HID) ButtonClosed(index uint8) bool {
	return d.Buttons[index].value
}

// Hat latest position.
// provided coords slice needs to be long enough to hold all the hat's axis.
func (d HID) HatCoords(index uint8, coords []float32) {
	for _, h := range d.HatAxes {
		if h.number == index {
			coords[h.axis-1] = h.value
		}
	}
	return
}

// insert events as if from hardware.
func (d HID) InsertSyntheticEvent(v int16, t uint8, i uint8) {
	d.OSEvents <- osEventRecord{Value: v, Type: t, Index: i}
}


