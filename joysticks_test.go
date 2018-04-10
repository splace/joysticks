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

func TestHIDsMutipleCapture(t *testing.T) {
	if !DeviceExists(1) && !DeviceExists(2) && !DeviceExists(3) && !DeviceExists(4){
		panic("no HIDs")
	}

	buttonEvents := Capture(
		Channel{10, HID.OnLong}, // button #10 long pressed
		Channel{1, HID.OnClose}, 
	)
	hatEvents := Capture(
		Channel{1, HID.OnHat},  
		Channel{2, HID.OnSpeedX},
		Channel{2, HID.OnPanX},
	)

	var x float32 = .5
	var f time.Duration = time.Second / 440
	for {
		select {
		case <-buttonEvents[0]:
			return
		case <-buttonEvents[1]:
			play(NewSound(NewTone(f, float64(x)), time.Second/3))
		case h := <-hatEvents[0]:
			f = time.Duration(100*math.Pow(2, float64(h.(CoordsEvent).Y))) * time.Second / 44000
		case h := <-hatEvents[1]:
			fmt.Printf("hat 2 X speed %+v\n",h.(AxisEvent).V)
		case h := <-hatEvents[2]:
			fmt.Printf("hat 2 X pan %+v\n",h.(AxisEvent).V)
		}
	}
}

func TestHIDsAdvanced(t *testing.T) {
	js1 := Connect(1)

	if js1 == nil {
		panic("no HID index 1")
	}
	if len(js1.Buttons) < 10 || len(js1.HatAxes) < 6 {
		t.Errorf("HID#1, available buttons %d, Hats %d\n", len(js1.Buttons), len(js1.HatAxes)/2)
	}

	b1 := js1.OnClose(1)
	b2 := js1.OnClose(2)
	b3 := js1.OnClose(3)
	b4 := js1.OnClose(4)
	DefaultRepeat=time.Second/10
	b5r := Repeater(js1.OnClose(5),js1.OnOpen(5))
	
	quit := js1.OnOpen(10)
	h3 :=  PositionFromVelocity(js1.OnMove(1))
	h4 := js1.OnPanX(2)
	h5 := js1.OnPanY(2)
	h6 := js1.OnEdge(1)
	h7 := js1.OnCenter(3)
	go js1.ParcelOutEvents()
	time.AfterFunc(time.Second*10, func() { js1.InsertSyntheticEvent(1, 1, 1) }) // value=1 (close),type=1 (button), index=1, so fires b1 after 10 seconds

	for {
		select {
		case <-quit:
			return
		case <-b1:
			DefaultRepeat=time.Second		
			play(NewSound(NewTone(time.Second/440, 1), time.Second/3))
		case <-b2:
			play(NewSound(NewTone(time.Second/660, 1), time.Second/3))
			b5r = Repeater(js1.OnClose(5),js1.OnOpen(5))
		case <-b3:
			play(NewSound(NewTone(time.Second/250, 1), time.Second/3))
		case <-b4:
			play(NewSound(NewTone(time.Second/150, 1), time.Second/3))
		case <-b5r:
			go play(NewSound(NewTone(time.Second/440, 1), time.Second/20))
		case h := <-h3:
			fmt.Printf("hat 1 moved %+v\n", h)
		case h := <-h4:
			fmt.Println("hat 2 X moved", h.(AxisEvent).V)
		case h := <-h5:
			fmt.Printf("hat 2 Y moved %+v\n", h)
		case h := <-h6:
			fmt.Println("hat 1 edged", h.(AngleEvent).Angle)
		case <-h7:
			fmt.Println("hat 3 centered")
		}
	}
}

func TestHIDsCapture(t *testing.T) {
	events := Capture(
		Channel{10, HID.OnDouble}, // event[0] button #10 double press
		Channel{1, HID.OnClose},   // event[1] button #1 closes
		Channel{1, HID.OnRotate},  // event[2] hat #1 rotates
		Channel{2, HID.OnRotate},  // event[2] hat #2 rotates
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
			fmt.Println(h.(AngleEvent).Angle)
			x = h.(AngleEvent).Angle/6.28 + .5
		case h := <-events[3]:
			fmt.Println(h.(AngleEvent).Angle)
			f = time.Duration(100*math.Pow(2, float64(h.(AngleEvent).Angle)/6.28)) * time.Second / 44000
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


