package regenbox

import (
	"errors"
	"log"
	"strconv"
	"sync"
	"time"
)

var ErrEmptyRead error = errors.New("message was empty")

//go:generate stringer -type=ChargeState
type ChargeState int

const (
	Idle        ChargeState = ChargeState(ModeIdle)
	Charging    ChargeState = ChargeState(ModeCharge)
	Discharging ChargeState = ChargeState(ModeDischarge)
)

//go:generate stringer -type=State
type State int

const (
	Disconnected    State = State(iota)
	Connected       State = State(iota)
	WriteError      State = State(iota)
	ReadError       State = State(iota)
	UnexpectedError State = State(iota)
	NilBox          State = State(iota)
)

//go:generate stringer -type=Mode
type Mode int

const (
	Manual        Mode = Mode(iota)
	DischargeOnly Mode = Mode(iota) // Discharge until BottomVoltage is reached, then idle
	ChargeOnly    Mode = Mode(iota) // Charge until TopVoltage is reached, then idle
	AutoRun       Mode = Mode(iota) // Do cycles up to NbCycles between Bottom & TopValues, then idle
)

type Snapshot struct {
	Time        time.Time
	Voltage     int
	ChargeState ChargeState
	State       State
}

type Config struct {
	OhmValue      int           // Value of charge resstance in ohm, usually from 10 to 30ohm
	Mode          Mode          // Auto-mode lets the box do charge cycles using the following config values
	NbHalfCycles  int           // In auto-mode: number of half-cycles to do before halting auto-mode
	UpDuration    time.Duration // In auto-mode: maximum time for an up-cycle before taking action
	DownDuration  time.Duration // In auto-mode: maximum time for a down-cycle before taking action
	TopVoltage    int           // In auto-mode: target top voltage before switching cycle
	BottomVoltage int           // In auto-mode: target bottom voltage before switching cycle
	IntervalSec   time.Duration // In auto-mode: sleep interval in second between each poll
}

type RegenBox struct {
	sync.Mutex
	Conn        *SerialConnection
	config      *Config
	chargeState ChargeState
	state       State
	stop        chan bool
	wg          sync.WaitGroup

	measures []Snapshot
}

func NewConfig() *Config {
	return &Config{
		OhmValue: 20, Mode: AutoRun, NbHalfCycles: 0,
		UpDuration: time.Hour * 12, DownDuration: time.Hour * 12,
		TopVoltage: 1450, BottomVoltage: 850,
		IntervalSec: time.Minute,
	}
}

func NewRegenBox(conn *SerialConnection, cfg *Config) (rb *RegenBox, err error) {
	if conn == nil {
		port, cfg, name, err := FindPort(nil)
		if err != nil {
			return nil, err
		}
		conn = NewSerial(port, cfg, name)
	}
	if cfg == nil {
		cfg = NewConfig()
	}

	rb = &RegenBox{
		Conn:        conn,
		config:      cfg,
		chargeState: Idle,
		state:       Connected,
	}

	_, err = rb.TestConnection()
	return rb, err
}

const (
	pingRetries  = 16
	testConnPoll = time.Millisecond * 250
)

// TestConnection sends a ping every testConnPoll,
// and returns on success or after pingRetries tries.
func (rb *RegenBox) TestConnection() (_ time.Duration, err error) {
	t0 := time.Now()
	for i := 0; i < pingRetries; i++ {
		time.Sleep(testConnPoll)
		err = rb.ping()
		if err == nil {
			break
		}
	}
	return time.Since(t0), err
}

// AutoRun starts an auto-cycle routine. To stop it, call StopAutoRun().
func (rb *RegenBox) AutoRun() {
	rb.stop = make(chan bool)
	rb.wg.Add(1)
	go func() {
		defer func() {
			rb.stop = nil // avoid closing of closed chan
			rb.wg.Done()

			log.Println("AutoRun is out, setting idle mode")
			err := rb.SetIdle()
			if err != nil {
				log.Println("in SetIdle():", err)
			}
		}()

		var sn Snapshot
		var halfCycles int
		var t0 = time.Now()
		for {
			select {
			case <-rb.stop:
				return
			case <-time.After(rb.config.IntervalSec):
			}

			// force charge state, can't hurt
			err := rb.SetChargeMode(byte(rb.chargeState))
			if err != nil {
				log.Println("in rb.SetChargeMode:", err)
			}

			sn = rb.Snapshot()
			log.Println(sn)
			rb.measures = append(rb.measures, sn)

			if sn.State != Connected {
				// need error-less state here
				continue
			}

			if rb.chargeState == Discharging {
				if sn.Voltage <= rb.config.BottomVoltage {
					log.Printf("bottom value %dmV reached", rb.config.BottomVoltage)
					if rb.config.Mode == DischargeOnly {
						log.Println("finished discharging battery (discharge only)")
						return
					}
					err := rb.SetCharge()
					if err != nil {
						log.Println("in rb.SetCharge:", err)
					} else {
						t0 = time.Now()
						halfCycles++
					}
				} else if time.Since(t0) >= rb.config.DownDuration {
					log.Printf("couldn't discharge battery to %dmV in %s, battery's dead or something's wrong",
						rb.config.BottomVoltage, rb.config.DownDuration)
					return
				}
			}

			if rb.chargeState == Charging {
				if sn.Voltage >= rb.config.TopVoltage {
					log.Printf("top value %dmV reached", rb.config.TopVoltage)
					if rb.config.Mode == ChargeOnly {
						log.Println("finished charging battery (charge only)")
						return
					}
					err := rb.SetDischarge()
					if err != nil {
						log.Println("in rb.SetDischarge:", err)
					} else {
						t0 = time.Now()
						halfCycles++
					}
				} else if time.Since(t0) >= rb.config.UpDuration {
					log.Printf("couldn't charge battery to %dmV in %s, battery's dead or something's wrong",
						rb.config.TopVoltage, rb.config.UpDuration)
					return
				}
			}

			if rb.config.NbHalfCycles > 0 && halfCycles >= rb.config.NbHalfCycles {
				log.Printf("reached target %d half-cycles", rb.config.NbHalfCycles)
				return
			}
		}
	}()
}

// StopAutoRun notifies AutoRun() to stop, and wait until it returns.
func (rb *RegenBox) StopAutoRun() {
	if rb.stop == nil {
		return
	}
	log.Println("stopping AutoRun...")
	close(rb.stop)
	rb.wg.Wait()
}

// Snapshot retreives the state of rb at a given time.
func (rb *RegenBox) Snapshot() Snapshot {
	s := Snapshot{
		Time:  time.Now(),
		State: rb.State(),
	}
	if s.State == NilBox {
		return s
	}
	var err error
	s.Voltage, err = rb.ReadVoltage()
	if err != nil {
		s.State = rb.state // update state, it should contain an error
		log.Printf("in rb.ReadVoltage: %s (state: %s)", err, s.State)
	}
	s.ChargeState = rb.ChargeState()
	return s
}

func (rb *RegenBox) Config() Config {
	return *rb.config
}

func (rb *RegenBox) SetConfig(cfg *Config) error {
	rb.config = cfg
	// todo check some values, take some actions, maybe
	return nil
}

func (rb *RegenBox) LedToggle() (bool, error) {
	rb.Lock()
	res, err := rb.talk(LedToggle)
	rb.Unlock()
	if err != nil {
		return false, err
	}
	if len(res) > 0 {
		return res[0] == 1, nil
	}
	return false, errors.New("empty read")

}

// ReadAnalog retreives value at A0 pin, it doesn't take
// account for CAN conversion. When in doubt, prefer ReadVoltage.
func (rb *RegenBox) ReadAnalog() (int, error) {
	rb.Lock()
	res, err := rb.talk(ReadA0)
	rb.Unlock()
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(string(res))
}

// ReadVoltage retreives voltage from battery on A0 in mV.
func (rb *RegenBox) ReadVoltage() (int, error) {
	rb.Lock()
	res, err := rb.talk(ReadVoltage)
	rb.Unlock()
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(string(res))
}

func (rb *RegenBox) SetCharge() error {
	return rb.SetChargeMode(ModeCharge)
}

func (rb *RegenBox) SetDischarge() error {
	return rb.SetChargeMode(ModeDischarge)
}

func (rb *RegenBox) SetIdle() error {
	return rb.SetChargeMode(ModeIdle)
}

func (rb *RegenBox) ChargeState() ChargeState {
	return rb.chargeState
}

func (rb *RegenBox) State() State {
	if rb == nil {
		return NilBox
	}
	return rb.state
}

// SetChargeMode sends mode instruction to regenbox.
func (rb *RegenBox) SetChargeMode(mode byte) error {
	rb.Lock()
	_, err := rb.talk(mode)
	rb.Unlock()
	if err != nil {
		return err
	}
	// no error, save state to box only now.
	rb.chargeState = ChargeState(mode)
	return nil
}

// ping sends a ping to regenbox, returning error if something's wrong
func (rb *RegenBox) ping() error {
	_, err := rb.talk(Ping)
	return err
}

// talk is generic 1-byte send and read []byte answer.
// All higher level function should use talk as a wrapper.
func (rb *RegenBox) talk(b byte) ([]byte, error) {
	i, err := rb.Conn.Write([]byte{b})
	if err != nil || i != 1 {
		rb.state = WriteError
		return nil, err
	}
	return rb.read()
}

// read reads from rb.Conn then returnes CRLF-trimmed response
func (rb *RegenBox) read() (buf []byte, err error) {
	buf = make([]byte, 256)
	i, err := rb.Conn.Read(buf)
	if err != nil {
		rb.state = ReadError
		return buf[:i], err
	}
	response := buf[:i]
	if i == 0 || len(response) == 0 {
		rb.state = UnexpectedError
		return []byte{}, ErrEmptyRead
	}
	rb.state = Connected
	return response, nil
}
