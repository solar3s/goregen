package regenbox

import (
	"errors"
	"github.com/rkjdid/util"
	"log"
	"strconv"
	"sync"
	"time"
)

var ErrDisconnected error = errors.New("no connection available")

var ErrBoxRunning = errors.New("box is already running")

var ErrCycleTimeout = errors.New("Timeout before reaching stop condition")
var ErrUserStop = errors.New("Stopped by user")
var ErrSnapshotSendTimeout = errors.New("Snapshot chan send timeout")

//go:generate stringer -type=State,ChargeState,BotMode -output=types_string.go
type State byte
type ChargeState byte
type BotMode byte

const (
	Idle        ChargeState = ChargeState(ModeIdle)
	Charging    ChargeState = ChargeState(ModeCharge)
	Discharging ChargeState = ChargeState(ModeDischarge)
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
	Charger    BotMode = BotMode(iota) // Charge until TopVoltage is reached, then idle
	Discharger                         // Discharge until BottomVoltage is reached, then idle
	Cycler                             // Do cycles up to NbCycles between Bottom & TopValues, then idle
)

type Snapshot struct {
	Time        time.Time
	Voltage     int
	ChargeState ChargeState
	State       State
	Firmware    string
}

type Config struct {
	Mode          BotMode       // Auto-mode lets the box do charge cycles using the following config values
	NbHalfCycles  int           // In auto-mode: number of half-cycles to do before halting auto-mode (0: no-limit holdem)
	UpDuration    util.Duration // In auto-mode: maximum time for an up-cycle before taking action (?)
	DownDuration  util.Duration // In auto-mode: maximum time for a down-cycle before taking action (?)
	TopVoltage    int           // In auto-mode: target top voltage before switching charge-cycle
	BottomVoltage int           // In auto-mode: target bottom voltage before switching charge-cycle
	Ticker        util.Duration // In auto-mode: sleep interval in second between each measure
	ChargeFirst   bool          // In auto-mode: start auto-run with a charge-cycle (false: discharge)
}

type RegenBox struct {
	sync.Mutex
	Conn        *SerialConnection
	config      *Config
	chargeState ChargeState
	state       State
	wg          sync.WaitGroup
	firmware    []byte

	stop     chan struct{}
	snapChan chan Snapshot
	msgChan  chan CycleMessage
}

var DefaultConfig = Config{
	Mode:          Charger,
	NbHalfCycles:  10,
	UpDuration:    util.Duration(time.Hour * 24),
	DownDuration:  util.Duration(time.Hour * 24),
	TopVoltage:    1500,
	BottomVoltage: 900,
	Ticker:        util.Duration(time.Second * 10),
	ChargeFirst:   false,
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
	} else if rb.firmware == nil {
		rb.firmware, _ = rb.talk(ReadFirmware)
	}
	return rb, err
}

const (
	pingRetries     = 30
	snapshotTimeout = time.Duration(time.Second * 5)
)

// TestConnection sends a ping every testConnPoll,
// and returns on success or after pingRetries tries.
func (rb *RegenBox) TestConnection() (_ time.Duration, err error) {
	rb.Lock()
	defer rb.Unlock()
	t0 := time.Now()
	for i := 0; i < pingRetries; i++ {
		err = rb.ping()
		if err == nil {
			rb.firmware, _ = rb.talk(ReadFirmware)
			break
		}
	}
	return time.Since(t0), err
}

func (rb *RegenBox) doCycle(tickerDuration util.Duration, maxDuration util.Duration, stopCond func(v int) bool) error {
	timeout := time.NewTimer(time.Duration(maxDuration))
	ticker := time.NewTicker(time.Duration(tickerDuration))
	var sn Snapshot
	for {
		select {
		case <-rb.stop:
			return ErrUserStop
		case <-timeout.C:
			return ErrCycleTimeout
		case <-ticker.C:
		}

		sn = rb.Snapshot()
		if sn.State != Connected {
			// need error-less state here, todo something?
			continue
		}

		// repeat charge state, just in case (e.g. usb connect drop)
		_ = rb.SetChargeMode(byte(rb.chargeState))

		// send snapshot through the pipe
		select {
		case rb.snapChan <- sn:
		case <-time.After(snapshotTimeout):
			return ErrSnapshotSendTimeout
		}

		if stopCond(sn.Voltage) {
			return nil
		}
	}
}

func (rb *RegenBox) topReached(i int) bool {
	return i >= rb.config.TopVoltage
}

func (rb *RegenBox) bottomReached(i int) bool {
	return i <= rb.config.BottomVoltage
}

// Start initiates a regen session, it returns a chan for snapshots,
// and a chan for end of cycles messages. If returned error is not nil, channels are nil.
func (rb *RegenBox) Start() (error, <-chan Snapshot, <-chan CycleMessage) {
	if !rb.Stopped() {
		return ErrBoxRunning, nil, nil
	}

	rb.snapChan = make(chan Snapshot)
	rb.msgChan = make(chan CycleMessage, 36)
	rb.stop = make(chan struct{})
	clean := func() {
		defer rb.wg.Done()
		rb.stop = nil
		var err error
		for i := 0; i < 3; i++ {
			err = rb.SetIdle()
			if err == nil {
				return
			}
			<-time.After(time.Millisecond * 250)
		}
		log.Println("error setting idle mode:", err)
	}

	var m CycleMessage
	switch rb.config.Mode {
	case Charger:
		err := rb.SetCharge()
		if err != nil {
			return err, nil, nil
		}

		rb.msgChan <- chargeStarted(rb.config.TopVoltage)
		rb.wg.Add(1)
		go func() {
			defer clean()
			err = rb.doCycle(rb.config.Ticker, rb.config.UpDuration, rb.topReached)
			if err == nil {
				m = chargeReached(rb.config.TopVoltage)
			} else if err == ErrCycleTimeout {
				m = chargeTimeout(rb.config.TopVoltage, rb.config.UpDuration)
			} else {
				m = chargeError(rb.config.TopVoltage, err)
			}
			rb.msgChan <- m
		}()
	case Discharger:
		err := rb.SetDischarge()
		if err != nil {
			return err, nil, nil
		}

		rb.msgChan <- dischargeStarted(rb.config.BottomVoltage)
		rb.wg.Add(1)
		go func() {
			defer clean()
			err = rb.doCycle(rb.config.Ticker, rb.config.DownDuration, rb.bottomReached)
			if err == nil {
				m = dischargeReached(rb.config.BottomVoltage)
			} else if err == ErrCycleTimeout {
				m = dischargeTimeout(rb.config.BottomVoltage, rb.config.DownDuration)
			} else {
				m = dischargeError(rb.config.BottomVoltage, err)
			}
			rb.msgChan <- m
		}()
	case Cycler:
		var err error
		if rb.config.ChargeFirst {
			err = rb.SetCharge()
		} else {
			err = rb.SetDischarge()
		}
		if err != nil {
			return err, nil, nil
		}

		rb.wg.Add(1)
		go func() {
			defer clean()

			var (
				err          error
				i, target    int
				duration     util.Duration
				currentCycle string
				nbCycles     = rb.config.NbHalfCycles
			)
			if rb.config.ChargeFirst {
				currentCycle = CycleDischarge
			}

			for i = 1; i <= nbCycles; i++ {
				if currentCycle == CycleDischarge {
					err = rb.SetCharge()
					if err != nil {
						break
					}
					currentCycle = CycleCharge
					target = rb.config.TopVoltage
					duration = rb.config.UpDuration
					rb.msgChan <- multiCycleStarted(target, currentCycle, i, nbCycles)
					err = rb.doCycle(rb.config.Ticker, duration, rb.topReached)
				} else {
					err = rb.SetDischarge()
					if err != nil {
						break
					}
					currentCycle = CycleDischarge
					target = rb.config.BottomVoltage
					duration = rb.config.DownDuration

					rb.msgChan <- multiCycleStarted(target, currentCycle, i, nbCycles)
					err = rb.doCycle(rb.config.Ticker, rb.config.DownDuration, rb.bottomReached)
				}
				if err != nil {
					break
				}
			}
			if err == nil {
				m = multiCycleReached(target, nbCycles)
			} else if err == ErrCycleTimeout {
				m = multiCycleTimeout(target, currentCycle, i, nbCycles, duration)
			} else {
				m = multiCycleError(target, err)
			}
			rb.msgChan <- m
		}()
	}

	return nil, rb.snapChan, rb.msgChan
}

// Stops the box, and wait until Start() loop returns.
func (rb *RegenBox) Stop() {
	if rb.Stopped() {
		return
	}
	close(rb.stop)
	rb.wg.Wait()
}

// Stopped returns false while box is running
func (rb *RegenBox) Stopped() bool {
	return rb.stop == nil
}

// Snapshot retreives the state of rb at a given time.
func (rb *RegenBox) Snapshot() Snapshot {
	s := Snapshot{
		Time:     time.Now(),
		State:    rb.State(),
		Firmware: string(rb.firmware),
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

func (rb *RegenBox) FirmwareVersion() string {
	return string(rb.firmware)
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
// /!\ This works because values match between
//    - ModeIdle/ModeCharge/ModeDischarge from protocol.go
//    - Idle/Charging/Discharging ChargeState from regenbox.go
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

// Ping sends a safe ping (locked) to regenbox
func (rb *RegenBox) Ping() error {
	rb.Lock()
	_, err := rb.talk(Ping)
	rb.Unlock()
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
	rb.state = Connected
	return buf, nil
}
