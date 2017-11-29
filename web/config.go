package web

import (
	"github.com/rkjdid/util"
	"github.com/solar3s/goregen/regenbox"
	"go.bug.st/serial.v1"
)

var DefaultConfig = Config{
	User:     NoName,
	Battery:  NoBattery,
	Resistor: 10,
	Web:      DefaultServerConfig,
	Regenbox: regenbox.DefaultConfig,
	Watcher:  regenbox.DefaultWatcherConfig,
	Serial:   regenbox.DefaultSerialConfig,
}

type Config struct {
	User     User
	Battery  Battery
	Resistor util.Float
	Regenbox regenbox.Config
	Web      ServerConfig
	Watcher  regenbox.WatcherConfig
	Device   string
	Serial   serial.Mode
}

type User struct {
	BetaId string // as provided by Regenbox
	Name   string
}

type Battery struct {
	BetaRef string // as provided by Regenbox
	Type    string // AAA, AA...
	Voltage int    // in millivolts
	Brand   string // Duracell...
	Model   string // Ultra
}

var NoBattery = Battery{}

var NoName = User{}
