package regenbox

import (
	"errors"
	"github.com/rkjdid/util"
	"log"
	"strconv"
	"sync"
	"time"
)

var ErrEmptyRead error = errors.New("message was empty")
var ErrDisconnected error = errors.New("no connection available")

//go:generate stringer -type=State,ChargeState,BotMode -output=types_string.go
type State byte
type ChargeState byte
type BotMode byte

const (
	Idle ChargeState = ChargeState(iota)
	Charging
	Discharging
)

const (
	Disconnected State = State(iota)
	Connected
	WriteError
	ReadError
	UnexpectedError
	NilBox
)

const (
	Manual     BotMode = BotMode(iota)
	Charger            // Charge until TopVoltage is reached, then idle
	Discharger         // Discharge until BottomVoltage is reached, then idle
	Cycler             // Do cycles up to NbCycles between Bottom & TopValues, then idle
)

type Snapshot struct {
	Time        time.Time
	Voltage     int
	ChargeState ChargeState
	State       State
}

type Config struct {
	Mode          BotMode       // Auto-mode lets the box do charge cycles using the following config values
	NbHalfCycles  int           // In auto-mode: number of half-cycles to do before halting auto-mode (0: no-limit holdem)
	UpDuration    util.Duration // In auto-mode: maximum time for an up-cycle before taking action (?)
	DownDuration  util.Duration // In auto-mode: maximum time for a down-cycle before taking action (?)
	TopVoltage    int           // In auto-mode: target top voltage before switching charge-cycle
	BottomVoltage int           // In auto-mode: target bottom voltage before switching charge-cycle
	IntervalSec   util.Duration // In auto-mode: sleep interval in second between each measure
	ChargeFirst   bool          // In auto-mode: start auto-run with a charge-cycle (false: discharge)
}

type RegenBox struct {
	sync.Mutex
	Conn        *SerialConnection
	config      *Config
	chargeState ChargeState
	state       State
	autorunCh   chan struct{}
	wg          sync.WaitGroup

	measures []Snapshot
}

var DefaultConfig = Config{
	Mode:          Charger,
	NbHalfCycles:  10,
	UpDuration:    util.Duration(time.Hour * 2),
	DownDuration:  util.Duration(time.Hour * 2),
	TopVoltage:    1410,
	BottomVoltage: 900,
	IntervalSec:   util.Duration(time.Second * 10),
	ChargeFirst:   true,
}

func NewConfig() *Config {
	var cfg = DefaultConfig
	return &cfg
}

func NewRegenBox(conn *SerialConnection, cfg *Config) (rb *RegenBox, err error) {
	if conn == nil {
		conn, err = FindSerial(nil)
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
	if conn == nil {
		rb.state = Disconnected
	}
	return rb, err
}

const (
	pingRetries = 12
)

// TestConnection sends a ping every testConnPoll,
// and returns on success or after pingRetries tries.
func (rb *RegenBox) TestConnection() (_ time.Duration, err error) {
	t0 := time.Now()
	for i := 0; i < pingRetries; i++ {
		err = rb.ping()
		if err == nil {
			break
		}
	}
	return time.Since(t0), err
}

// Starts a detached routine. To stop it, call StopAutoRun()
func (rb *RegenBox) Start() {
	logChargeState := func(i int) {
		log.Printf("autorun: %s (%d)", rb.chargeState, i)
	}

	rb.autorunCh = make(chan struct{})
	rb.wg.Add(1)
	if (rb.config.Mode == Charger) || rb.config.ChargeFirst {
		rb.chargeState = Charging
	} else {
		rb.chargeState = Discharging
	}
	logChargeState(0)
	go func() {
		defer func() {
			rb.autorunCh = nil // avoid closing of closed chan
			rb.wg.Done()

			log.Println("autorun out, setting idle mode")
			err := rb.SetIdle()
			if err != nil {
				log.Println("in SetIdle():", err)
			}
		}()

		var sn Snapshot
		var halfCycles int = 1
		var t0 = time.Now()
		for {
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
					log.Printf("autorun: %dV reached bottom value", rb.config.BottomVoltage)
					if rb.config.Mode == Discharger {
						log.Println("finished discharging battery (discharge only)")
						return
					}
					err := rb.SetCharge()
					if err != nil {
						log.Println("in rb.SetCharge:", err)
					} else {
						logChargeState(halfCycles)
						t0 = time.Now()
						halfCycles++
					}
				} else if time.Since(t0) >= time.Duration(rb.config.DownDuration) {
					log.Printf("autorun: couldn't discharge battery to %dmV in %s, battery's dead or something's wrong",
						rb.config.BottomVoltage, rb.config.DownDuration)
					return
				}
			}

			if rb.chargeState == Charging {
				if sn.Voltage >= rb.config.TopVoltage {
					log.Printf("autorun: %dV reached top limit", rb.config.TopVoltage)
					if rb.config.Mode == Charger {
						log.Println("finished charging battery (charge only)")
						return
					}
					err := rb.SetDischarge()
					if err != nil {
						log.Println("in rb.SetDischarge:", err)
					} else {
						logChargeState(halfCycles)
						t0 = time.Now()
						halfCycles++
					}
				} else if time.Since(t0) >= time.Duration(rb.config.UpDuration) {
					log.Printf("couldn't charge battery to %dmV in %s, battery's dead or something's wrong",
						rb.config.TopVoltage, rb.config.UpDuration)
					return
				}
			}

			if rb.config.NbHalfCycles > 0 && halfCycles >= rb.config.NbHalfCycles {
				log.Printf("reached target %d half-cycles", rb.config.NbHalfCycles)
				return
			}

			select {
			case <-rb.autorunCh:
				return
			case <-time.After(time.Duration(rb.config.IntervalSec)):
			}
		}
	}()
}

// Stops the box, and wait until Start() loop returns.
func (rb *RegenBox) Stop() {
	if rb.Stopped() {
		return
	}
	log.Println("stopping AutoRun...")
	close(rb.autorunCh)
	rb.wg.Wait()
}

// Stopped returns false while box is running
func (rb *RegenBox) Stopped() bool {
	return rb.autorunCh == nil
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
// TODO THIS IS SHIT cause ModeIdle... Idle/Charging/Discharging are not compatible
// TODO at least test this t_t
func (rb *RegenBox) SetChargeMode(charge byte) error {
	var mode byte
	switch charge {
	case byte(Idle):
		mode = byte(ModeIdle)
	case byte(Charging):
		mode = byte(ModeCharge)
	case byte(Discharging):
		mode = byte(ModeDischarge)
	default:
		mode = byte(charge)
	}

	switch mode {
	case ModeIdle:
		charge = byte(Idle)
	case ModeCharge:
		charge = byte(Charging)
	case ModeDischarge:
		charge = byte(Discharging)
	default:
		charge = byte(Idle)
		fmt.Println("wut?", mode)
	}

	rb.Lock()
	_, err := rb.talk(mode)
	rb.Unlock()
	if err != nil {
		return err
	}
	// no error, save state to box only now.
	rb.chargeState = ChargeState(charge)
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
	if rb.Conn == nil || rb.state == Disconnected {
		return nil, ErrDisconnected
	}
	err := rb.Conn.Write([]byte{b})
	if err != nil {
		rb.state = WriteError
		return nil, err
	}
	return rb.read()
}

// read reads from rb.Conn then returnes CRLF-trimmed response
func (rb *RegenBox) read() (buf []byte, err error) {
	buf, err = rb.Conn.Read()
	if err != nil {
		rb.state = ReadError
		return buf, err
	}
	if buf == nil || len(buf) == 0 {
		rb.state = UnexpectedError
		return []byte{}, ErrEmptyRead
	}
	rb.state = Connected
	return buf, nil
}
