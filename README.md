# joysticks

Go language joystick/controller/gamepad input.

uses Linux kernel 'input' interface, available on a wide range of linux devices, to receive events directly, no polling.

uses go channels to pipe around events, for flexibility and multi-threading.

# Operation

make event channels from methods on HID type.  'HID.On***()'  

also some higher-level event modifiers for common UI abstractions, to help standard usage.

Overview/docs: [![GoDoc](https://godoc.org/github.com/splace/joysticks?status.svg)](https://godoc.org/github.com/splace/joysticks)

# Installation:

     go get github.com/splace/joysticks

# Examples: 

highlevel: block until button one pressed.

	package main

	import . "github.com/splace/joysticks"

	func main() {
		evts := Capture(
			Channel{1, HID.OnClose},  // event[0] chan set to receive button #1 closes events
		)
		<-evts[0]
	}


print out description of event when pressing button #1 or moving hat#1.(with 10sec timeout.) 

	package main

	import . "github.com/splace/joysticks"
	import "fmt"
	import  "time"

	func main() {
		device := Connect(1)

		if device == nil {
			panic("no HIDs")
		}

		// using Connect allows a device to be interrogated
		fmt.Printf("HID#1:- Buttons:%d, Hats:%d\n", len(device.Buttons), len(device.HatAxes)/2)

		// make channels for specific events
		b1press := device.OnClose(1)
		h1move := device.OnMove(1)

		// feed OS events onto the event channels. 
		go device.ParcelOutEvents()

		// handle event channels
		go func(){
			for{
				select {
				case <-b1press:
					fmt.Println("button #1 pressed")
				case h := <-h1move:
					hpos:=h.(CoordsEvent)
					fmt.Println("hat #1 moved too:", hpos.X,hpos.Y)
				}
			}
		}()
	
		fmt.Println("Timeout in 10 secs.")
		time.Sleep(time.Second*10)
		fmt.Println("Shutting down due to timeout.")
	}



Note: "jstest-gtk" - gtk mapping and calibration for joysticks.


