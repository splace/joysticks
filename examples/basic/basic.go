package main

//import "log"
//import "time"
import "os"
import . "github.com/splace/joysticks"
import "fmt"

func main() {
	deviceReader, err := os.OpenFile("/dev/input/js1", os.O_RDWR, 0)
	var OSEvents chan osEventRecord
	go eventPipe(deviceReader, OSEvents)
	
}

