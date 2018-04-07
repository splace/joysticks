package main

import . "github.com/splace/joysticks"

func main() {
	events := Capture(
		Channel{1, HID.OnClose},  // event[0] is button #1 closes
	)
	<-events[0]
}

