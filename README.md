# joysticks
Go language joystick interface.

uses channels for events, so multi-threaded.

Overview/docs: [![GoDoc](https://godoc.org/github.com/splace/joysticks?status.svg)](https://godoc.org/github.com/splace/joysticks)

Installation:

     go get github.com/splace/joysticks

Example: play a note when pressing button #1. hat position changes frequency, y axis, and volume, x axis. (double press button #10 to exit) 

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
			Channel{10, State.OnLong}, // events[0] button #10 long pressed
			Channel{1, State.OnClose}, // events[1] button #1 closes
			Channel{1, State.OnMove},  // events[2] hat #1 moves
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




