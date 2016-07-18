package joysticks

import (
	"fmt"
	"testing"
)
import . "github.com/splace/sounds"

import (
	"io"
	"os/exec"
	"time"
)
import "math"

func TestHIDsCapture(t *testing.T) {
	events := Capture(
		Channel{10, HID.OnLong},  // event[0] button #10 long pressed
		Channel{1, HID.OnClose},  // event[1] button #1 closes
		Channel{1, HID.OnRotate}, // event[2] hat #1 rotates
		Channel{2, HID.OnRotate}, // event[2] hat #1 rotates
	)
	var x float32 = .5
	var f time.Duration = time.Second / 440
	for {
		select {
		case <-events[0]:
			return
		case <-events[1]:
			play(NewSound(NewTone(f, float64(x)), time.Second/3))
		case h := <-events[2]:
			fmt.Println(h.(HatAngleEvent).Angle)
			x = h.(HatAngleEvent).Angle/6.28 + .5
		case h := <-events[3]:
			fmt.Println(h.(HatAngleEvent).Angle)
			f = time.Duration(100*math.Pow(2, float64(h.(HatAngleEvent).Angle)/6.28)) * time.Second / 44000
		}
	}
}

func TestHIDsMutipleCapture(t *testing.T) {
	events1 := Capture(
		Channel{10, HID.OnLong}, // event[0] button #10 long pressed
		Channel{1, HID.OnClose}, // event[1] button #1 closes
		Channel{1, HID.OnMove},  // event[2] hat #1 moves
	)
	events2 := Capture(
		Channel{10, HID.OnLong}, // event[0] button #10 long pressed
		Channel{1, HID.OnClose}, // event[1] button #1 closes
		Channel{1, HID.OnMove},  // event[2] hat #1 moves
	)
	var x float32 = .5
	var f time.Duration = time.Second / 440
	for {
		select {
		case <-events1[0]:
			return
		case <-events2[0]:
			return
		case <-events1[1]:
			play(NewSound(NewTone(f, float64(x)), time.Second/3))
		case <-events2[1]:
			play(NewSound(NewTone(f, float64(x)), time.Second/3))
		case h := <-events1[2]:
			x = h.(HatPositionEvent).X/2 + .5
			f = time.Duration(100*math.Pow(2, float64(h.(HatPositionEvent).Y))) * time.Second / 44000
		case h := <-events2[2]:
			x = h.(HatPositionEvent).X/2 + .5
			f = time.Duration(100*math.Pow(2, float64(h.(HatPositionEvent).Y))) * time.Second / 44000
		}
	}
}

func TestHIDsAdvanced(t *testing.T) {
	js1 := Connect(1)

	if js1 == nil {
		panic("no HIDs")
	}
	if len(js1.Buttons) < 10 || len(js1.HatAxes) < 6 {
		t.Errorf("HID#1, available buttons %d, Hats %d\n", len(js1.Buttons), len(js1.HatAxes)/2)
	}

	b1 := js1.OnClose(1)
	b2 := js1.OnClose(2)
	b3 := js1.OnClose(3)
	b4 := js1.OnClose(4)
	quit := js1.OnOpen(10)
	h1 := js1.OnMove(1)
	h4 := js1.OnPanX(2)
	h5 := js1.OnPanY(2)
	h3 := js1.OnMove(3)
	go js1.ParcelOutEvents()
	time.AfterFunc(time.Second*10, func() { js1.InsertSyntheticEvent(1, 1, 1) }) // value=1 (close),type=1 (button), index=1, so fires b1 after 10 seconds

	for {
		select {
		case <-quit:
			return
		case <-b1:
			play(NewSound(NewTone(time.Second/440, 1), time.Second/3))
		case <-b2:
			play(NewSound(NewTone(time.Second/660, 1), time.Second/3))
		case <-b3:
			play(NewSound(NewTone(time.Second/250, 1), time.Second/3))
		case <-b4:
			play(NewSound(NewTone(time.Second/150, 1), time.Second/3))
		case h := <-h1:
			fmt.Println("hat 1 moved", h)
		case h := <-h3:
			fmt.Println("hat 3 moved", h)
		case h := <-h4:
			fmt.Println("hat 2 X moved", h.(HatPanXEvent).V)
		case h := <-h5:
			fmt.Println("hat 2 Y moved", h)
		}
	}
}

func play(s Sound) {
	out, in := io.Pipe()
	go func() {
		Encode(in, 2, 44100, s)
		in.Close()
	}()
	cmd := exec.Command("aplay")
	cmd.Stdin = out
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}
