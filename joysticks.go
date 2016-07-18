package joysticks

import (
	"math"
	"time"
)
var LongPressDelay = time.Second

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

//HID holds the in-coming event channel, mappings, and registered events for a device, and has methods to control and adjust behaviour.
type HID struct {
	OSEvents              chan osEventRecord
	Buttons               map[uint8]button
	HatAxes               map[uint8]hatAxis
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

// Hat Axis changed event, X,Y {-1...1}
type HatPositionEvent struct {
	when
	X, Y float32
}

// Button changed event
type ButtonChangeEvent struct {
	when
}

// Hat Axis changed event, V {-1...1}
type HatPanXEvent struct {
	when
	V float32
}

// Hat Axis changed event, V {-1...1}
type HatPanYEvent struct {
	when
	V float32
}

// Hat angle changed event, Angle {-Pi...Pi}
type HatAngleEvent struct {
	when
	Angle float32
}

// ParcelOutEvents waits on the HID.OSEvent channel (so is blocking), then puts the required event(s), on any registered channel(s).
func (d HID) ParcelOutEvents() {
	for {
		if evt, ok := <-d.OSEvents; ok {
			switch evt.Type {
			case 1:
				b := d.Buttons[evt.Index]
				if evt.Value == 0 {
					if c, ok := d.buttonOpenEvents[b.number]; ok {
						c <- ButtonChangeEvent{when{toDuration(evt.Time)}}
					}
					if c, ok := d.buttonLongPressEvents[b.number]; ok {
						if toDuration(evt.Time) > b.time+LongPressDelay {
							c <- ButtonChangeEvent{when{toDuration(evt.Time)}}
						}
					}
				}
				if evt.Value == 1 {
					if c, ok := d.buttonCloseEvents[b.number]; ok {
						c <- ButtonChangeEvent{when{toDuration(evt.Time)}}
					}
				}
				d.Buttons[evt.Index] = button{b.number, toDuration(evt.Time), evt.Value != 0}
			case 2:
				h := d.HatAxes[evt.Index]
				v := float32(evt.Value) / maxValue
				switch h.axis {
				case 1:
					if c, ok := d.hatPanXEvents[h.number]; ok {
						c <- HatPanXEvent{when{toDuration(evt.Time)}, v}
					}
				case 2:
					if c, ok := d.hatPanYEvents[h.number]; ok {
						c <- HatPanYEvent{when{toDuration(evt.Time)}, v}
					}
				}
				if c, ok := d.hatPositionEvents[h.number]; ok {
					switch d.HatAxes[evt.Index].axis {
					case 1:
						c <- HatPositionEvent{when{toDuration(evt.Time)}, v, d.HatAxes[evt.Index+1].value}
					case 2:
						c <- HatPositionEvent{when{toDuration(evt.Time)}, d.HatAxes[evt.Index-1].value, v}
					}
				}
				if c, ok := d.hatAngleEvents[h.number]; ok {
					switch d.HatAxes[evt.Index].axis {
					case 1:
						c <- HatAngleEvent{when{toDuration(evt.Time)}, float32(math.Atan2(float64(v), float64(d.HatAxes[evt.Index+1].value)))}
					case 2:
						c <- HatAngleEvent{when{toDuration(evt.Time)}, float32(math.Atan2(float64(d.HatAxes[evt.Index-1].value), float64(v)))}
					}
				}
				d.HatAxes[evt.Index] = hatAxis{h.number, h.axis, toDuration(evt.Time), v}
			default:
				// log.Println("unknown input type. ",evt.Type & 0x7f)
			}
		} else {
			break
		}
	}
}

// Type of registerable methods and the index they are called witd. (Note: the event type is indicated by the method.)
type Channel struct {
	Number uint8
	Method func(HID, uint8) chan event
}

// button goes open
func (d HID) OnOpen(button uint8) chan event {
	c := make(chan event)
	d.buttonOpenEvents[button] = c
	return c
}

// button goes closed
func (d HID) OnClose(button uint8) chan event {
	c := make(chan event)
	d.buttonCloseEvents[button] = c
	return c
}

// button goes open and last event on it, closed, wasn't recent. (within 1 second)
func (d HID) OnLong(button uint8) chan event {
	c := make(chan event)
	d.buttonLongPressEvents[button] = c
	return c
}

// hat moved
func (d HID) OnMove(hat uint8) chan event {
	c := make(chan event)
	d.hatPositionEvents[hat] = c
	return c
}

// hat axis-X moved
func (d HID) OnPanX(hat uint8) chan event {
	c := make(chan event)
	d.hatPanXEvents[hat] = c
	return c
}

// hat axis-Y moved
func (d HID) OnPanY(hat uint8) chan event {
	c := make(chan event)
	d.hatPanYEvents[hat] = c
	return c
}

// hat angle changed
func (d HID) OnRotate(hat uint8) chan event {
	c := make(chan event)
	d.hatAngleEvents[hat] = c
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

// insert events as if from hardware.
func (d HID) InsertSyntheticEvent(v int16, t uint8, i uint8) {
	d.OSEvents <- osEventRecord{Value: v, Type: t, Index: i}
}
