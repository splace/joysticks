package main

import . "github.com/splace/joysticks"

func main() {
	evts := Capture(
		Channel{1, HID.OnClose},  // event[0] chan set to receive button #1 closes events
	)
	<-evts[0]
}

