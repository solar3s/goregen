package main

import (
	"flag"
	"fmt"
	"github.com/rkjdid/util"
	"github.com/solar3s/goregen/regenbox"
	"github.com/solar3s/goregen/web"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

var (
	conn       *regenbox.SerialConnection
	rootConfig *web.Config
)

var (
	device     = flag.String("dev", "-", "path to serial port, if empty it will be searched automatically")
	rootPath   = flag.String("root", "", "path to goregen's main directory (defaults to executable path)")
	cfgPath    = flag.String("config", "", "path to config (defaults to <root>/config.toml)")
	assetsPath = flag.String("assets", "", "restore static assets to provided directory & exit")
	logDir     = flag.String("log", "", "path to logs directory (defaults to <root>/log)")
	dataDir    = flag.String("data", "", "path to data directory (defaults to <root>/data)")
	verbose    = flag.Bool("v", false, "higher verbosity")
	version    = flag.Bool("version", false, "print version & exit")
)

func init() {
	flag.Parse()

	// print version & exit
	if *version {
		fmt.Printf("goregen %s\n", Version)
		os.Exit(0)
	}

	// restore static assets & exit
	/*if *assetsPath != "" {
		err := web.RestoreAssets(*assetsPath, "static")
		if err != nil {
			log.Fatalf("couldn't restore static assets in \"%s\": %s", *assetsPath, err)
		} else {
			p, _ := filepath.Abs(*assetsPath)
			log.Printf("restored assets to directory \"%s\"", filepath.Join(p, "static"))
			log.Println("use it as Web.StaticDir value in config to prioritize extracted static assets")
			os.Exit(0)
		}
	}*/

	// root directory for goregen
	if *rootPath == "" {
		exe, err := os.Executable()
		if err != nil {
			log.Fatalf("couldn't get path to executable: %s", err)
		}
		*rootPath = filepath.Dir(exe)
	}

	err := os.MkdirAll(*rootPath, 0755)
	if err != nil {
		log.Fatalf("couldn't mkdir root directory \"%s\": %s", *rootPath, err)
	}

	// create log file
	if *logDir == "" {
		*logDir = filepath.Join(*rootPath, "log")
	}
	err = os.MkdirAll(*logDir, 0755)
	if err != nil {
		log.Fatalf("couldn't mkdir log directory \"%s\": %s", *logDir, err)
	}

	logPath := filepath.Join(*logDir, time.Now().Format("2006-01-02_15h04m05.log"))
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("couldn't create log file: %s", err)
	}

	// create log link
	link := "goregen.log"
	logLink := filepath.Join(*rootPath, link)
	_ = os.Remove(logLink)
	err = os.Symlink(logPath, logLink)
	if err != nil {
		err = os.Link(logPath, logLink)
		if err != nil {
			log.Fatalf("couldn't create \"%s\" link: %s", link, err)
		}
	}

	// log to both Stderr & logFile
	log.SetOutput(io.MultiWriter(logFile, os.Stderr))

	// load config
	if *cfgPath == "" {
		*cfgPath = filepath.Join(*rootPath, "config.toml")
	}
	err = util.ReadTomlFile(&rootConfig, *cfgPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("error reading config \"%s\": %s", *cfgPath, err)
		}
		rootConfig = &web.DefaultConfig
		err = util.WriteTomlFile(rootConfig, *cfgPath)
		if err != nil {
			log.Fatalf("error creating config file \"%s\": %s", *cfgPath, err)
		}
		log.Printf("created new config file \"%s\"", *cfgPath)
	}

	// override device from -dev flag
	if *device != "-" {
		rootConfig.Device = *device
	}
	// explicit device address
	if rootConfig.Device != "" {
		dev := rootConfig.Device
		port, config, err := regenbox.OpenPortName(dev)
		if err != nil {
			log.Fatal("error opening serial port: ", err)
		}
		conn = regenbox.NewSerial(port, config, dev, true)
		conn.Start()
	}

	// datalogs directory
	if *dataDir != "" {
		rootConfig.Web.DataDir = *dataDir
	}
	err = os.MkdirAll(rootConfig.Web.DataDir, 0755)
	if err != nil {
		log.Fatalf("couldn't mkdir Web.DataDir \"%s\": %s", rootConfig.Web.DataDir, err)
	}

	log.Printf("using config file: %s", *cfgPath)

	// write full config for future references in log file only
	_ = util.WriteToml(rootConfig, logFile)
}

func main() {
	rbox, err := regenbox.NewRegenBox(conn, &rootConfig.Regenbox)
	if err != nil {
		log.Println("error scanning for RegenBox:", err)
	}
	if conn != nil {
		_, err := rbox.TestConnection()
		if err != nil {
			log.Printf("no response from regenbox on port \"%s\": %s", *device, err)
			os.Exit(1)
		} else {
			log.Printf("connected to \"%s\"", *device)
		}
	}

	log.Printf("starting conn watcher (poll rate: %s)", rootConfig.Watcher.ConnPollRate)
	watcher := regenbox.NewWatcher(rbox, &rootConfig.Watcher)
	watcher.WatchConn()

	log.Printf("starting webserver on http://%s ...", rootConfig.Web.ListenAddr)
	go web.StartServer(Version, rbox, rootConfig, *cfgPath, *verbose)

	// small delay to allow for panic in StartServer
	<-time.After(time.Millisecond * 500)
	log.Println("Press <Ctrl-C> to quit")

	trap := make(chan os.Signal)
	signal.Notify(trap, os.Kill, os.Interrupt)
	<-trap
	fmt.Println()
	log.Println("quit received...")

	cleanExit := make(chan struct{})
	go func() {
		watcher.Stop()
		rbox.Stop()
		if rbox.Conn != nil {
			rbox.Conn.Close()
		}

		close(cleanExit)
	}()
	select {
	case <-time.After(time.Second * 10):
		log.Panicln("no clean exit after 10sec, please report panic log to https://github.com/solar3s/goregen/issues")
	case <-cleanExit:
	}
}
