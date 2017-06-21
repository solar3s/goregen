package web

import (
	"github.com/solar3s/goregen/regenbox"
	"go.bug.st/serial.v1"
)

var DefaultConfig = Config{
	User:     NoName,
	Battery:  NoBattery,
	Web:      DefaultServerConfig,
	Regenbox: regenbox.DefaultConfig,
	Watcher:  regenbox.DefaultWatcherConfig,
	Serial:   regenbox.DefaultSerialConfig,
}

type Config struct {
	User     User
	Battery  Battery
	Web      ServerConfig
	Regenbox regenbox.Config
	Watcher  regenbox.WatcherConfig
	Serial   serial.Mode
}

type User struct {
	Name string
}

type Battery struct {
	Type    string // AAA, AA...
	Voltage int    // in millivolts
	Brand   string // Duracell...
	Model   string // Ultra
}

var NoBattery = Battery{}

var NoName = User{}
