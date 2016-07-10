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

func TestJoysticksCapture(t *testing.T) {
	events := Capture(
		Channel{10, state.OnOpen}, // event[0] button #10 opens
		Channel{1, state.OnClose}, // event[1] button #1 closes
		Channel{1, state.OnMove},  // event[2] hat #1 moves
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
			x = h.(HatChangeEvent).X/2 + .5
			f = time.Duration(100*math.Pow(2, float64(h.(HatChangeEvent).Y))) * time.Second / 44000
		}
	}
}

func TestJoysticksAdvanced(t *testing.T) {
	js1, err := Connect(1)

	if err != nil {
		panic(err)
	}
	if len(js1.buttons) < 10 || len(js1.hatAxes) < 6 {
		t.Errorf("joystick#1, available buttons %d, Hats %d\n", len(js1.buttons), len(js1.hatAxes)/2)
	}

	b1 := js1.OnClose(1)
	b2 := js1.OnClose(2)
	b3 := js1.OnClose(3)
	b4 := js1.OnClose(4)
	quit := js1.OnOpen(10)
	h1 := js1.OnMove(1)
	h2 := js1.OnMove(2)
	h3 := js1.OnMove(3)
	go js1.ProcessEvents()
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
		case h := <-h2:
			fmt.Println("hat 2 moved", h)
		case h := <-h3:
			fmt.Println("hat 3 moved", h)
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


