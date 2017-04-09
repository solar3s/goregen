package main

import (
	"flag"
	"github.com/solar3s/goregen/regenbox"
	"go.bug.st/serial.v1"
	"log"
	"time"
)

var rbox *regenbox.RegenBox
var err error

var (
	device = flag.String("dev", "", "path to serial port, if empty it will be searched automatically")
)

func init() {
	flag.Parse()
	var conn regenbox.Connection = nil
	if *device != "" {
		port, err := serial.Open(*device, regenbox.DefaultSerialMode)
		if err != nil {
			log.Fatal("error opening serial port:", err)
		}
		conn = regenbox.SerialConnection{Port: port}
	}
	rbox, err = regenbox.NewRegenBox(conn)
	if err != nil {
		log.Fatal("error initializing regenbox connection:", err)
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
