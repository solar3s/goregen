package regenbox

import (
	"errors"
	"go.bug.st/serial.v1"
	"log"
)

var ErrNoSerialPortFound = errors.New("didn't find any available serial port")

var DefaultSerialMode = &serial.Mode{
	BaudRate: 9600,
	Parity:   serial.NoParity,
	DataBits: 8,
	StopBits: serial.OneStopBit,
}

type SerialConnection struct {
	serial.Port
}

func (sc SerialConnection) Read(b []byte) (int, error) {
	return sc.Port.Read(b)
}

func (sc SerialConnection) Write(b []byte) (int, error) {
	return sc.Port.Write(b)
}

func (sc SerialConnection) Close() error {
	return sc.Port.Close()
}

// AutoConnectSerial connects automatically to
// the first successfully opened serial port (platform independant hopefully).
// If mode is nil, DefaultSerialMode is used.
func AutoConnectSerial(mode *serial.Mode) (Connection, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}
	if mode == nil {
		mode = DefaultSerialMode
	}
	var port serial.Port
	for _, v := range ports {
		port, err = serial.Open(v, mode)
		if err == nil {
			log.Printf("Found serial port \"%s\"", v)
			log.Printf("Options: %#v", mode)
			return &SerialConnection{Port: port}, nil
		}
	}
	if err == nil {
		return nil, ErrNoSerialPortFound
	}
	return nil, err
}
