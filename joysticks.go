package joysticks

import (
	"math"
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

//HID holds the in-coming event channel, mappings, and registered events for a joystick, and has methods to control and adjust behaviour.
type HID struct {
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

// ParcelOutEvents interprets waits on the HID.OSEvent channel (so is blocking), then puts the required event(s), on any registered channel(s).
func (h HID) ParcelOutEvents() {
	for {
		if evt, ok := <-h.OSEvent; ok {
			switch evt.Type {
			case 1:
				if evt.Value == 0 {
					if c, ok := h.buttonOpenEvents[h.buttons[evt.Index].number]; ok {
						c <- ButtonChangeEvent{when{toDuration(evt.Time)}}
					}
					if c, ok := h.buttonLongPressEvents[h.buttons[evt.Index].number]; ok {
						if toDuration(evt.Time) > h.buttons[evt.Index].time+time.Second {
							c <- ButtonChangeEvent{when{toDuration(evt.Time)}}
						}
					}
				}
				if evt.Value == 1 {
					if c, ok := h.buttonCloseEvents[h.buttons[evt.Index].number]; ok {
						c <- ButtonChangeEvent{when{toDuration(evt.Time)}}
					}
				}
				h.buttons[evt.Index] = button{h.buttons[evt.Index].number, toDuration(evt.Time), evt.Value != 0}
			case 2:
				switch h.hatAxes[evt.Index].axis {
				case 1:
					if c, ok := h.hatPanXEvents[h.hatAxes[evt.Index].number]; ok {
						c <- HatPanXEvent{when{toDuration(evt.Time)}, float32(evt.Value) / maxValue}
					}
				case 2:
					if c, ok := h.hatPanYEvents[h.hatAxes[evt.Index].number]; ok {
						c <- HatPanYEvent{when{toDuration(evt.Time)}, float32(evt.Value) / maxValue}
					}
				}
				if c, ok := h.hatPositionEvents[h.hatAxes[evt.Index].number]; ok {
					switch h.hatAxes[evt.Index].axis {
					case 1:
						c <- HatPositionEvent{when{toDuration(evt.Time)}, float32(evt.Value) / maxValue, h.hatAxes[evt.Index+1].value}
					case 2:
						c <- HatPositionEvent{when{toDuration(evt.Time)}, h.hatAxes[evt.Index-1].value, float32(evt.Value) / maxValue}
					}
				}
				if c, ok := h.hatAngleEvents[h.hatAxes[evt.Index].number]; ok {
					switch h.hatAxes[evt.Index].axis {
					case 1:
						c <- HatAngleEvent{when{toDuration(evt.Time)}, float32(math.Atan2(float64(evt.Value), float64(h.hatAxes[evt.Index+1].value)))}
					case 2:
						c <- HatAngleEvent{when{toDuration(evt.Time)}, float32(math.Atan2(float64(h.hatAxes[evt.Index-1].value), float64(evt.Value)/maxValue))}
					}
				}
				h.hatAxes[evt.Index] = hatAxis{h.hatAxes[evt.Index].number, h.hatAxes[evt.Index].axis, toDuration(evt.Time), float32(evt.Value) / maxValue}
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
	Method func(HID, uint8) chan event
}

// button goes open
func (h HID) OnOpen(button uint8) chan event {
	c := make(chan event)
	h.buttonOpenEvents[button] = c
	return c
}

// button goes closed
func (h HID) OnClose(button uint8) chan event {
	c := make(chan event)
	h.buttonCloseEvents[button] = c
	return c
}

// button goes open and last event on it, closed, wasn't recent. (within 1 second)
func (h HID) OnLong(button uint8) chan event {
	c := make(chan event)
	h.buttonLongPressEvents[button] = c
	return c
}

// hat moved
func (h HID) OnMove(hat uint8) chan event {
	c := make(chan event)
	h.hatPositionEvents[hat] = c
	return c
}

// hat axis-X moved
func (h HID) OnPanX(hat uint8) chan event {
	c := make(chan event)
	h.hatPanXEvents[hat] = c
	return c
}

// hat axis-Y moved
func (h HID) OnPanY(hat uint8) chan event {
	c := make(chan event)
	h.hatPanYEvents[hat] = c
	return c
}

// hat axis-Y moved
func (h HID) OnRotate(hat uint8) chan event {
	c := make(chan event)
	h.hatAngleEvents[hat] = c
	return c
}

func (h HID) ButtonExists(button uint8) (ok bool) {
	for _, v := range h.buttons {
		if v.number == button {
			return true
		}
	}
	return
}

func (h HID) HatExists(hat uint8) (ok bool) {
	for _, v := range h.hatAxes {
		if v.number == hat {
			return true
		}
	}
	return
}

func (h HID) InsertSyntheticEvent(v int16, t uint8, i uint8) {
	h.OSEvent <- osEventRecord{Value: v, Type: t, Index: i}
}
