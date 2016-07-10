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
		Channel{10, State.OnLong}, // event[0] button #10 double pressed
		Channel{1, State.OnClose}, // event[1] button #1 closes
		Channel{1, State.OnMove},  // event[2] hat #1 moves
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


