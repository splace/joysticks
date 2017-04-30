package joysticks

import (
	"time"
	//"fmt"
)

// TODO drag event
// TODO move plus edge continue events (self generating)
// TODO smoother from PID
// TODO two pans to a coord
// TODO coords to two pans
// TODO 1-d integrator

var DefaultRepeat = time.Second /4
var VelocityRepeat =  time.Second / 10


// duplicate event onto two chan's
func Duplicator(c chan Event)(chan Event,chan Event){
	c1 := make(chan Event)
	c2 := make(chan Event)
	go func(){
		for e:=range c{
			c1 <- e
			c2 <- e
		}
		close(c1)
		close(c2)
	}()
	return c1,c2
}


// creates a chan on which you get CoordsEvent's that are the time integration of the CoordsEvent's on the parameter chan. 
func PositionFromVelocity(c chan Event) chan Event{
	extra := make(chan Event)
	var x,y,vx,vy float32
	var startTime time.Time
	var startMoment time.Duration
	var m time.Duration
	var lt time.Time
	ticker:=time.NewTicker(VelocityRepeat)
	// receiving chan processor
	go func(){
		e:= <-c
		startTime=time.Now()
		startMoment=e.Moment()
		lm:=startMoment
		for e:=range c{
			if ce,ok:=e.(CoordsEvent);ok{
				lt=time.Now()
				m=e.Moment()
				dt:=float32((m-lm).Seconds())
				x+=vx*dt
				y+=vy*dt
				vx,vy=ce.X,ce.Y
				lm=	m
			}
		}
		ticker.Stop()
	}()
	// output chan processor
	go func(){
		var lx,ly,nx,ny,dt float32
		for t:=range ticker.C{
			dt=float32(t.Sub(lt).Seconds())
			nx,ny=x+dt*vx,y+dt*vy
			if nx!=lx || ny!=ly {
				extra <-CoordsEvent{when{startMoment+t.Sub(startTime)},nx,ny}	
				lx,ly=nx,ny
			}
		}
	}()

	return extra
}


// creates a channel that, after receiving any event on the first parameter chan, and until any event on second chan parameter, regularly receives 'when' events.
// the repeat interval is DefaultRepeat, and is stored, so retriggering is not effected by changing DefaultRepeat.
func Repeater(c1,c2 chan Event)(chan Event){
	c := make(chan Event)
	go func(){
		interval:=DefaultRepeat
		var ticker *time.Ticker
		for {
			e:= <-c1
			go func(interval time.Duration, startTime time.Time){
				ticker=time.NewTicker(interval)
				for t:=range ticker.C{
					c <- when{e.Moment()+t.Sub(startTime)}
				}
			}(interval, time.Now())
			<-c2
			ticker.Stop()
		}
	}()
	return c
}

