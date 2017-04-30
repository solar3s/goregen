package main

import (
	"flag"
	"github.com/solar3s/goregen/regenbox"
	"github.com/solar3s/goregen/www"
	"log"
	"time"
)

var conn *regenbox.SerialConnection
var server *www.Server

var (
	device  = flag.String("dev", "", "path to serial port, if empty it will be searched automatically")
	verbose = flag.Bool("v", false, "higher verbosity")
)

func init() {
	flag.Parse()
	if *device != "" {
		port, config, err := regenbox.OpenPortName(*device)
		if err != nil {
			log.Fatal("error opening serial port: ", err)
		}
		conn = regenbox.NewSerial(port, config)
	}
}

func main() {
	cfg := &regenbox.Config{
		OhmValue:      20,
		Mode:          regenbox.ChargeOnly,
		NbHalfCycles:  0,
		UpDuration:    time.Hour * 6,
		DownDuration:  time.Hour * 6,
		TopVoltage:    1500,
		BottomVoltage: 850,
		IntervalSec:   time.Second * 10,
	}
	rbox, err := regenbox.NewRegenBox(conn, cfg)
	if err != nil {
		log.Fatal("error initializing regenbox connection: ", err)
	}

	log.Println("connected to", rbox.Conn.Path())
	server = &www.Server{
		ListenAddr: "localhost:3636",
		Regenbox:   rbox,
		Verbose:    *verbose,
	}
	err = server.Start()
	if err != nil {
		log.Fatal(err)
	}
}
