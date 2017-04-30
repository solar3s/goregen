package regenbox

import (
	"errors"
	"go.bug.st/serial.v1"
	"log"
)

var ErrNoSerialPortFound = errors.New("didn't find any available serial port")

var DefaultSerialConfig = &serial.Mode{
	BaudRate: 9600,
	Parity:   serial.NoParity,
	DataBits: 8,
	StopBits: serial.OneStopBit,
}

type SerialConnection struct {
	serial.Port
	path   string
	config *serial.Mode
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
func FindPort(config *serial.Mode) (serial.Port, *serial.Mode, string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, nil, "", err
	}
	if config == nil {
		config = DefaultSerialConfig
	}
	var port serial.Port
	for _, v := range ports {
		port, err = serial.Open(v, config)
		if err == nil {
			log.Printf("found serial port \"%s\"", v)
			return port, config, v, nil
		}
	}
	if err == nil {
		return nil, config, "", ErrNoSerialPortFound
	}
	return nil, config, "", err
}

func OpenPortName(name string) (port serial.Port, config *serial.Mode, err error) {
	config = DefaultSerialConfig
	port, err = serial.Open(name, config)
	return port, config, err
}

func NewSerial(port serial.Port, config *serial.Mode, name string) *SerialConnection {
	return &SerialConnection{
		Port:   port,
		path:   name,
		config: config,
	}
}
