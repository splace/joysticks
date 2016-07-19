package main

import (
	"io"
	"os/exec"
	"time"
	"math"
)

import . "github.com/splace/joysticks"

import . "github.com/splace/sounds"

func main() {
	events := Capture(
		Channel{10, HID.OnLong},  // event[0] button #10 long pressed
		Channel{1, HID.OnClose},  // event[1] button #1 closes
		Channel{1, HID.OnRotate}, // event[2] hat #1 rotates
		Channel{2, HID.OnRotate}, // event[2] hat #2 rotates
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
			x = h.(HatAngleEvent).Angle/6.28 + .5
		case h := <-events[3]:
			f = time.Duration(100*math.Pow(2, float64(h.(HatAngleEvent).Angle)/6.28)) * time.Second / 44000
		}
	}
}

func play(s Sound) {
	cmd := exec.Command("aplay")
	out, in := io.Pipe()
	go func() {
		Encode(in, 2, 44100, s)
		in.Close()
	}()
	cmd.Stdin = out
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
} 



