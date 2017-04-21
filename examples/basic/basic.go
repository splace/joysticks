package main

import . "github.com/splace/joysticks"
import "fmt"
import  "time"

func main() {
	device := Connect(1)

	if device == nil {
		panic("no HIDs")
	}
	fmt.Printf("HID#1:- Buttons:%d, Hats:%d\n", len(device.Buttons), len(device.HatAxes)/2)

	// make channels for specific events
	b1press := device.OnClose(1)
	b2press := device.OnClose(2)
	h1move := device.OnMove(1)

	// feed OS events onto the event channels. 
	go device.ParcelOutEvents()

	// handle event channels
	go func(){
		for{
			select {
			case <-b1press:
				fmt.Println("button #1 pressed")
			case <-b2press:
				fmt.Println("button #2 pressed")
				coords := make([]float32, 2)
				device.ReadHatPosition(1,coords)
				fmt.Println("current hat #1 position:",coords)
			case h := <-h1move:
				hpos:=h.(HatPositionEvent)
				fmt.Println("hat #1 moved too:", hpos.X,hpos.Y)
			}
		}
	}()
	
	fmt.Println("Timeout in 10 secs.")
	<-time.After(time.Second*10)
	fmt.Println("Shutting down due to timeout.")
}

