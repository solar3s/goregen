package regenbox

import (
	"errors"
	"fmt"
)

var ErrEmptyRead error = errors.New("message was empty")

type RegenBox struct {
	Conn Connection
}

func NewRegenBox(conn Connection) (rb *RegenBox, err error) {
	if conn == nil {
		conn, err = AutoConnectSerial(nil)
		if err != nil {
			return nil, err
		}
	}

	rb = &RegenBox{conn}
	buf := make([]byte, 1)
	_, err = rb.Conn.Read(buf)
	if err != nil {
		return rb, err
	} else if buf[0] != BoxReady {
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

func (rb *RegenBox) ReadVoltage() (string, error) {
	res, err := rb.talk(ReadVoltage)
	if err != nil {
		return "", err
	}
	return string(res), err
}

// talk is generic 1-byte send and read []byte answer
func (rb *RegenBox) talk(b byte) ([]byte, error) {
	i, err := rb.Conn.Write([]byte{b})
	if err != nil || i != 1 {
		return nil, err
	}
	return rb.readTrim()
}

// readTrim reads all it can get from rb.Conn, then removes CRLF
func (rb *RegenBox) readTrim() ([]byte, error) {
	buf := make([]byte, 256)
	i, err := rb.Conn.Read(buf)
	if err != nil {
		return buf[:i], err
	}
	out := trimCRLF(buf[:i])
	if i == 0 || len(out) == 0 {
		return []byte{}, ErrEmptyRead
	}
	return out, nil
}
