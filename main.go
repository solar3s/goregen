package main

import (
	"flag"
	"github.com/solar3s/goregen/regenbox"
	"github.com/solar3s/goregen/www"
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

func dumbCycle() {
	log.Println("enabling discharge")
	err := rbox.SetDischarge()
	if err != nil {
		log.Println("rbox.SetDischarge error:", err)
	}

	up := false
	for {
		time.Sleep(time.Minute)
		rbox.LedToggle()
		rV, err := rbox.ReadVoltage()
		if err != nil {
			log.Println("err ReadRoltage:", err)
		}

		log.Printf("Voltage: %vmV", rV)
		if rV < 900 {
			up = true
		} else if rV >= 1400 {
			up = false
		}

		if up {
			err := rbox.SetCharge()
			if err != nil {
				log.Println("rbox.SetCharge error:", err)
			}
		} else {
			err := rbox.SetDischarge()
			if err != nil {
				log.Println("rbox.SetDischarge error:", err)
			}
		}
	}
}

func main() {
	go dumbCycle() // run our dumb cycle in background

	s := www.Server{
		ListenAddr: "localhost:3636",
		Regenbox:   rbox,
	}
	err := s.Start()
	if err != nil {
		log.Fatal(err)
	}
}
