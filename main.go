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
	debug   = flag.Bool("debug", false, "enable debug mode")
)

func init() {
	flag.Parse()
	if *device != "" {
		port, config, err := regenbox.OpenPortName(*device)
		if err != nil {
			log.Fatal("error opening serial port: ", err)
		}
		conn = regenbox.NewSerial(port, config, *device)
		conn.Start()
	}
}

func main() {
	cfg := &regenbox.Config{
		OhmValue:      20,
		Mode:          regenbox.ChargeOnly,
		NbHalfCycles:  10,
		UpDuration:    time.Hour * 2,
		DownDuration:  time.Hour * 2,
		TopVoltage:    1410,
		BottomVoltage: 900,
		IntervalSec:   time.Second * 10,
		ChargeFirst:   true,
	}
	rbox, err := regenbox.NewRegenBox(conn, cfg)
	if err != nil {
		log.Println("error initializing regenbox connection:", err)
	}

	server = &www.Server{
		ListenAddr: "localhost:3636",
		Regenbox:   rbox,
		Verbose:    *verbose,
		Debug:      *debug,
	}
	server.Start()
	<-make(chan bool)
}
