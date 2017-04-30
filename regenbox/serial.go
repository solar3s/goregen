package regenbox

import (
	"errors"
	"github.com/tarm/serial"
	port_discover "go.bug.st/serial.v1"
	"log"
	"time"
)

var ErrNoSerialPortFound = errors.New("didn't find any available serial port")

var DefaultSerialConfig = &serial.Config{
	Baud:        9600,
	Parity:      serial.ParityNone,
	Size:        serial.DefaultSize,
	StopBits:    serial.Stop1,
	ReadTimeout: time.Millisecond * 500,
}

type SerialConnection struct {
	*serial.Port
	path   string
	config *serial.Config
}

func (sc *SerialConnection) Read(b []byte) (int, error) {
	return sc.Port.Read(b)
}

func (sc *SerialConnection) Write(b []byte) (int, error) {
	return sc.Port.Write(b)
}

func (sc *SerialConnection) Close() error {
	return sc.Port.Close()
}

func (sc *SerialConnection) Path() string {
	return sc.path
}

// FindPort tries to connect to first available serial port (platform independant hopefully).
// If mode is nil, DefaultSerialMode is used.
func FindPort(config *serial.Config) (*serial.Port, *serial.Config, error) {
	ports, err := port_discover.GetPortsList()
	if err != nil {
		return nil, nil, err
	}
	if config == nil {
		config = DefaultSerialConfig
	}
	var port *serial.Port
	for _, v := range ports {
		config.Name = v
		port, err = serial.OpenPort(config)
		if err == nil {
			log.Printf("found serial port \"%s\"", v)
			return port, config, nil
		}
	}
	if err == nil {
		return nil, config, ErrNoSerialPortFound
	}
	return nil, config, err
}

func OpenPortName(name string) (port *serial.Port, config *serial.Config, err error) {
	config = DefaultSerialConfig
	config.Name = name
	port, err = serial.OpenPort(config)
	return port, config, err
}

func NewSerial(port *serial.Port, config *serial.Config) *SerialConnection {
	return &SerialConnection{
		Port:   port,
		path:   config.Name,
		config: config,
	}
}
