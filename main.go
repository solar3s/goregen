package main

import (
	"flag"
	"fmt"
	"github.com/rkjdid/util"
	"github.com/solar3s/goregen/regenbox"
	"github.com/solar3s/goregen/web"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"time"
)

var (
	conn       *regenbox.SerialConnection
	rootConfig *web.Config
	staticPath string
)

var (
	device     = flag.String("dev", "", "path to serial port, if empty it will be searched automatically")
	rootPath   = flag.String("root", "", "path to goregen's main directory (defaults to executable path)")
	cfgPath    = flag.String("config", "", "path to config (defaults to <root>/config.toml)")
	verbose    = flag.Bool("verbose", false, "higher verbosity")
	version    = flag.Bool("version", false, "print version & exit")
	assetsPath = flag.Bool("assets", false, "extract static assets to <root>/static, if true, extracted assets "+
		"also take precedence over binary assets. This option is useful for doing live tests on front-end")
)

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func init() {
	flag.Parse()
	if *version {
		fmt.Printf("goregen %s\n", Version)
		os.Exit(0)
	}

	if *device != "" {
		port, config, err := regenbox.OpenPortName(*device)
		if err != nil {
			log.Fatal("error opening serial port: ", err)
		}
		conn = regenbox.NewSerial(port, config, *device)
		conn.Start()
	}

	if *rootPath == "" {
		exe, err := os.Executable()
		if err != nil {
			log.Fatalf("couldn't get path to executable: %s", err)
		}
		*rootPath = filepath.Dir(exe)
	}
	for _, v := range []string{*rootPath} {
		err := os.MkdirAll(v, 0755)
		if err != nil {
			log.Fatalf("couldn't mkdir \"%s\": %s", v, err)
		}
	}

	if *cfgPath == "" {
		*cfgPath = filepath.Join(*rootPath, "config.toml")
	}

	err := util.ReadTomlFile(&rootConfig, *cfgPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("error reading config \"%s\": %s", *cfgPath, err)
		}
		rootConfig = &web.DefaultConfig
		err = util.WriteTomlFile(rootConfig, *cfgPath)
		if err != nil {
			log.Fatalf("error creating config \"%s\": %s", *cfgPath, err)
		}
		log.Printf("created new config file \"%s\"", *cfgPath)
	}

	// restore static assets
	if *assetsPath {
		err = web.RestoreAssets(*rootPath, "static")
		if err != nil {
			log.Fatalf("couldn't restore static assets in \"%s\": %s", staticPath, err)
		} else {
			rootConfig.Web.StaticDir = filepath.Join(*rootPath, "static")
			log.Printf("restored static assets to: %s", rootConfig.Web.StaticDir)
		}
	}

	if *verbose {
		rootConfig.Web.Verbose = true
	}

	log.Printf("using config file: %s", *cfgPath)
}

func main() {
	rbox, err := regenbox.NewRegenBox(conn, &rootConfig.Regenbox)
	if err != nil {
		log.Println("error initializing regenbox connection:", err)
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
	server := web.NewServer(Version, rbox, rootConfig)
	server.Start()

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
