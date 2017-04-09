package main

import (
	"flag"
	"github.com/tarm/serial"
	"log"
	"time"
	"github.com/solar3s/regenbox/goregen/regenbox"
)

var rbox *regenbox.RegenBox
var err error

var (
	device = flag.String("dev", "/dev/ttyUSB1", "path to serial port")
)

func init() {
	flag.Parse()
	dev, err := serial.OpenPort(&serial.Config{
		Name:        *device,
		Baud:        9600,
		ReadTimeout: time.Millisecond * 500,
	})
	if err != nil {
		log.Fatal(err)
	}
	rbox = &regenbox.RegenBox{
		Conn: regenbox.SerialConnection{
			Port: dev,
		},
	}
}

// ReadVoltage on A0 pin
func main() {
	for {
		time.Sleep(time.Second)
		r0, err := rbox.LedToggle()
		if err != nil {
			log.Println(err)
		}
		r1, err := rbox.ReadVoltage()
		if err != nil {
			log.Println(err)
		}
		log.Println("led:", r0, " - A0:", r1)
	}
}
