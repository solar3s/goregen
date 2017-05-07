package regenbox

import (
	"errors"
	"fmt"
	"go.bug.st/serial.v1"
	"log"
	"sync"
	"time"
)

var ErrNoSerialPortFound = errors.New("didn't find any available serial port")
var ErrClosedPort = errors.New("serial port is closed")

var DefaultSerialConfig = &serial.Mode{
	BaudRate: 57600,
	Parity:   serial.NoParity,
	DataBits: 8,
	StopBits: serial.OneStopBit,
}

var DefaultTimeout = time.Second

type SerialConnection struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	serial.Port
	path   string
	config *serial.Mode

	rdChan    chan []byte
	wrChan    chan []byte
	errChan   chan error
	closeChan chan struct{}
	wg        sync.WaitGroup
}

func NewSerial(port serial.Port, config *serial.Mode, name string) *SerialConnection {
	return &SerialConnection{
		Port:      port,
		path:      name,
		config:    config,
		rdChan:    make(chan []byte),
		wrChan:    make(chan []byte),
		errChan:   make(chan error),
		closeChan: make(chan struct{}),

		ReadTimeout:  DefaultTimeout,
		WriteTimeout: DefaultTimeout,
	}
}

// Start begins the two routines responsible
// for reading and writing on serial port.
func (sc *SerialConnection) Start() {
	sc.wg.Add(2)
	go func() {
		sc.readRoutine()
		sc.wg.Done()
	}()
	go func() {
		sc.writeRoutine()
		sc.wg.Done()
	}()
}

// Read takes one of sc.rdChan or sc.errChan, waiting up to sc.ReadTimeout,
// it also checks if connection is closed and returns error accordingly.
func (sc *SerialConnection) Read() (b []byte, err error) {
	select {
	case b = <-sc.rdChan:
	case err = <-sc.errChan:
	case <-sc.closeChan:
		err = ErrClosedPort
	case <-time.After(sc.ReadTimeout):
		err = fmt.Errorf("read timeout (%s)", sc.ReadTimeout)
	}
	return b, err
}

// Write pushes b to sc.wrChan, or returns an error
// after sc.WriteTimeout, or if connection is closed.
func (sc *SerialConnection) Write(b []byte) (err error) {
	select {
	case sc.wrChan <- b:
	case <-sc.closeChan:
		err = ErrClosedPort
	case <-time.After(sc.WriteTimeout):
		err = fmt.Errorf("write timeout (%s)", sc.WriteTimeout)
	}
	return err
}

// Close notifies read/write routines to stop, then waits
// for them to return, it then actually closes serial port.
func (sc *SerialConnection) Close() error {
	close(sc.closeChan)
	sc.wg.Wait()
	return sc.Port.Close()
}

// Path returns device name / path of serial port.
func (sc *SerialConnection) Path() string {
	return sc.path
}

func (sc *SerialConnection) readRoutine() {
	for {
		time.Sleep(time.Millisecond * 50)
		b := make([]byte, 32)
		i, err := sc.Port.Read(b)
		if err != nil {
			select {
			case sc.errChan <- err:
			case <-sc.closeChan:
				return
			}
		} else {
			select {
			case sc.rdChan <- b[:i]:
			case <-sc.closeChan:
				return
			}
		}
	}
}

func (sc *SerialConnection) writeRoutine() {
	var b []byte
	for {
		time.Sleep(time.Millisecond * 50)
		select {
		case b = <-sc.wrChan:
		case <-sc.closeChan:
			return
		}
		_, err := sc.Port.Write(b)
		if err != nil {
			log.Println("in sc.writeRoutine:", err)
		}
	}
}

// FindSerial tries to connect to first available serial port (platform independant hopefully).
// If config is nil, DefaultSerialMode is used.
func FindSerial(config *serial.Mode) (*SerialConnection, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}
	if config == nil {
		config = DefaultSerialConfig
	}
	var port serial.Port
	for _, v := range ports {
		port, err = serial.Open(v, config)
		if err == nil {
			log.Printf("trying \"%s\"...", v)
			conn := NewSerial(port, config, v)
			conn.ReadTimeout = time.Millisecond * 250
			conn.WriteTimeout = time.Millisecond * 250
			conn.Start()
			// create a temporary box to test connection
			rb := &RegenBox{Conn: conn, config: new(Config), state: Connected}
			t, err := rb.TestConnection()
			if err == nil {
				log.Printf("connected to \"%s\" in %s", v, t)
				return conn, nil
			}
		}
	}
	if err == nil {
		return nil, ErrNoSerialPortFound
	}
	return nil, err
}

func OpenPortName(name string) (port serial.Port, config *serial.Mode, err error) {
	config = DefaultSerialConfig
	port, err = serial.Open(name, config)
	return port, config, err
}
