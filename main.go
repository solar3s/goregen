package main

import (
	"flag"
	"github.com/solar3s/goregen/regenbox"
	"github.com/solar3s/goregen/www"
	"go.bug.st/serial.v1"
	"log"
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
	rbox, err = regenbox.NewRegenBox(conn, nil)
	if err != nil {
		log.Fatal("error initializing regenbox connection:", err)
	}
}

func main() {
	s := www.Server{
		ListenAddr: "localhost:3636",
		Regenbox:   rbox,
	}
	err := s.Start()
	if err != nil {
		log.Fatal(err)
	}
}
