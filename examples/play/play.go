package main

import "log"
import "time"
import . "github.com/splace/joysticks"

func main() {
	timer:=time.After(time.Second*10)
	log.Println("press any button to start.")
	events := Capture(
		Channel{1, HID.OnClose},  // event[0] button #1 closes
	)
	log.Println("started with 10 second timeout.")
loop:
	for {
		select {
	    case <-timer:
			log.Println("Shutting down server due to timeout.")
			break loop
		case <-events[0]:
			log.Println("Button #1 pressed")
		}
	}
}

