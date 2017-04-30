package main

import "github.com/splace/joysticks"
import "log"
import  "time"

type Region struct{
	minX,minY,maxX,maxY float32
}

type RegionEvent struct {
	time time.Duration
	Index int
}

func (b RegionEvent) Moment() time.Duration {
	return b.time
}


func OnRegionChanged(trigger,position chan joysticks.Event, regions []Region) (chan RegionEvent,chan RegionEvent){
	in := make(chan RegionEvent)
	out := make(chan RegionEvent)
	states:=make([]bool,len(regions))
	var pevt joysticks.CoordsEvent
	go func(){
		for e:=range(position){
			switch v := e.(type) {
			case joysticks.CoordsEvent:
				pevt=v
		    }
		}
	}()
	go func(){
		for e:=range(trigger){
			for ri,r:=range(regions){
				if pevt.X>=r.minX && pevt.X<=r.maxX && pevt.Y>=r.minY && pevt.Y<=r.maxY{ 
					if !states[ri]{
						in <-RegionEvent{e.Moment(),ri}
						states[ri]=true
					}
				}else{
					if states[ri]{
						out <-RegionEvent{e.Moment(),ri}
						states[ri]=false
					}
				}
			}
		}
		close(in)
		close(out)
	}()
	return in,out
}


func main() {
	device := joysticks.Connect(1)
	if device == nil {
		log.Println("no HIDs")
		return
	}
	log.Printf("HID#1:- Buttons:%d, Hats:%d\n", len(device.Buttons), len(device.HatAxes)/2)

	// make channels for specific events
	regionEntering,regionExiting:=OnRegionChanged(device.OnHat(1),device.OnMove(1),[]Region{Region{.5,.5,1,1},Region{-1,-1,-.5,-.5}})

	// feed OS events onto the event channels. 
	go device.ParcelOutEvents()

	// handle event channels
	go func(){
		for{
			select {
			case r:= <-regionEntering:
				log.Println("enter region:",r.Index)
			case r:= <-regionExiting:
				log.Println("leave region:",r.Index)
			}
		}
	}()
	
	log.Println("Timeout in 10 secs.")
	<-time.After(time.Second*10)
	log.Println("Shutting down due to timeout.")
}


