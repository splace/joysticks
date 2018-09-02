# joysticks

Go language joystick/controller/gamepad input.

uses Linux kernel 'input' interface, available on a wide range of linux devices, to receive events directly, no polling.

uses go channels to pipe around events, for flexibility and multi-threading.

## Operation

make event channels from methods on HID type.  `HID.On<<event>>()`

also some higher-level event modifiers for common UI abstractions, to help standard usage.

Overview/docs: [![GoDoc](https://godoc.org/github.com/splace/joysticks?status.svg)](https://godoc.org/github.com/splace/joysticks)

## Installation:

     go get github.com/splace/joysticks

## Examples: 

### highlevel 

automates Connection to an available device, event chan creation and parcelling out those events 

```` Go
// block until button one pressed.
package main

import . "github.com/splace/joysticks"

func main() {
	evts := Capture(
		Channel{1, HID.OnClose},  // chan index#0 set to receive button #1 closes events
	)
	<-evts[0]
}
````

### Midlevel 

allows device interrogation and event re-assigning.

```` Go
// log a description of events when pressing button #1 or moving hat#1. 
// 10sec timeout.
package main

import . "github.com/splace/joysticks"
import "log"
import "time"

func main() {
	// try connecting to specific controller.
	// the index is system assigned, typically it increments on each new controller added.
	// indexes remain fixed for a given controller, if/when other controller(s) are removed.
	device := Connect(1)

	if device == nil {
		panic("no HIDs")
	}

	// using Connect allows a device to be interrogated
	log.Printf("HID#1:- Buttons:%d, Hats:%d\n", len(device.Buttons), len(device.HatAxes)/2)

	// get/assign channels for specific events
	b1press := device.OnClose(1)
	h1move := device.OnMove(1)

	// start feeding OS events onto the event channels. 
	go device.ParcelOutEvents()

	// handle event channels
	go func(){
		for{
			select {
			case <-b1press:
				log.Println("button #1 pressed")
			case h := <-h1move:
				hpos:=h.(CoordsEvent)
				log.Println("hat #1 moved too:", hpos.X,hpos.Y)
			}
		}
	}()

	log.Println("Timeout in 10 secs.")
	time.Sleep(time.Second*10)
	log.Println("Shutting down due to timeout.")
}
````

Note: "jstest-gtk" - gtk mapping and calibration for joysticks.


