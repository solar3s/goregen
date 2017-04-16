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

// forever cycle
func main() {
	log.Println("enabling discharge")
	err := rbox.SetDischarge()
	if err != nil {
		log.Fatal("couldn't set dischargecharge mode", err)
	}

	for {
		time.Sleep(time.Second)
		rbox.LedToggle()
		rV, err := rbox.ReadVoltage()
		if err != nil {
			log.Println("err ReadRoltage:", err)
		}

		log.Printf("Voltage: %vmV", rV)
		if rV < 900 && rbox.ChargeState() != regenbox.Charging {
			log.Printf("current state: %v. enabling charge", rbox.ChargeState())
			err := rbox.SetCharge()
			if err != nil {
				log.Println("rbox.SetCharge error:", err)
			}
		} else if rV >= 1400 && rbox.ChargeState() != regenbox.Discharging {
			log.Printf("current state: %v. enabling discharge", rbox.ChargeState())
			err := rbox.SetDischarge()
			if err != nil {
				log.Println("rbox.SetDischarge error:", err)
			}
		}
	}
}
