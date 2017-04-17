package regenbox

import (
	"errors"
	"fmt"
	"strconv"
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

type RegenBox struct {
	Conn        Connection
	chargeState ChargeState
	state       State
}

func NewRegenBox(conn Connection) (rb *RegenBox, err error) {
	if conn == nil {
		conn, err = AutoConnectSerial(nil)
		if err != nil {
			return nil, err
		}
	}

	rb = &RegenBox{conn, Idle, Connected}
	buf, err := rb.read()
	if err != nil {
		return rb, err
	} else if buf[0] != BoxReady {
		rb.state = UnexpectedError
		return rb, fmt.Errorf("unexpected ready byte %s vs. %d", buf[0], BoxReady)
	}
	return rb, nil
}

func (rb *RegenBox) LedToggle() (bool, error) {
	res, err := rb.talk(LedToggle)
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
	res, err := rb.talk(ReadA0)
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(string(res))
}

// ReadVoltage retreives voltage from battery on A0 in mV.
func (rb *RegenBox) ReadVoltage() (int, error) {
	res, err := rb.talk(ReadVoltage)
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

// SetDischarge enables discharge mode
func (rb *RegenBox) SetChargeMode(mode byte) error {
	res, err := rb.talk(mode)
	if err != nil {
		return err
	}
	if len(res) < 1 {
		return errors.New("no response from regenbox")
	}
	// no error, save state to box only now.
	rb.chargeState = ChargeState(mode)
	return nil
}

// talk is generic 1-byte send and read []byte answer.
// All higher level function should use talk as a wrapper.
func (rb *RegenBox) talk(b byte) ([]byte, error) {
	time.Sleep(time.Millisecond * 50) // small tempo
	i, err := rb.Conn.Write([]byte{b})
	if err != nil || i != 1 {
		rb.state = WriteError
		return nil, err
	}
	return rb.read()
}

// Read reads from rb.Conn then removes CRLF
func (rb *RegenBox) read() (buf []byte, err error) {
	buf = make([]byte, 256)
	i, err := rb.Conn.Read(buf)
	if err != nil {
		rb.state = ReadError
		return buf[:i], err
	}
	out := trimCRLF(buf[:i])
	if i == 0 || len(out) == 0 {
		rb.state = UnexpectedError
		return []byte{}, ErrEmptyRead
	}
	rb.state = Connected
	return out, nil
}
