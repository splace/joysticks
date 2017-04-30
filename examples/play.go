package main

import (
	"io"
	"os/exec"
	"time"
	"math"
	"os"
	"os/signal"
	"log"
)

import . "github.com/splace/joysticks"

import . "github.com/splace/sounds"


func main() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	events := Capture(
		Channel{10, HID.OnLong},  // event[0] button #10 long pressed
		Channel{1, HID.OnClose},  // event[1] button #1 closes
		Channel{7, HID.OnOpen},  // event[2] button #7 opens
		Channel{8, HID.OnOpen},  // event[3] button #8 opens
		Channel{2, HID.OnRotate}, // event[4] hat #1 rotates
		Channel{1, HID.OnEdge}, // event[5] hat #2 rotates
	)
	var volume float32 = .5
	var octave =5
	var note int =1
	for {
		select {
		case <-stopChan: // wait for SIGINT
			log.Println("Interrupted")
			return
		case <-events[0]:
			return
		case <-events[1]:
			play(NewSound(NewTone(Period(octave,note), float64(volume)), time.Second/3))
		case <-events[2]:
			octave++
		case <-events[3]:
			octave--
		case h := <-events[4]:
			volume = h.(HatAngleEvent).Angle/6.28 + .5
		case h := <-events[5]:
			//f = time.Duration(100*math.Pow(2, float64(h.(HatAngleEvent).Angle)/6.28)) * time.Second / 44000
			note = int(h.(HatAngleEvent).Angle*6 / math.Pi)+6
			play(NewSound(NewTone(Period(octave,note), float64(volume)), time.Second/3))
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
	log.Println("Playing:",s)
} 



