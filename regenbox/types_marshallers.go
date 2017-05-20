package regenbox

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// This file contains (un)marshallers for various byte types
// used in regenbox, allowing to more easily encode / decode
// string values instead of byte values, making communication
// with any front-end or config files easier.
//
// this file should be go-generated, too

// ---- type State int

func (s State) MarshalJSON() ([]byte, error) {
	b, err := s.MarshalText()
	if err == nil {
		b = []byte(fmt.Sprintf("\"%s\"", string(b)))
	}
	return b, err
}

func (s *State) UnmarshalJSON(data []byte) error {
	dataLength := len(data)
	if data[0] != '"' || data[dataLength-1] != '"' {
		return errors.New("State.UnmarshalJSON: Invalid JSON provided")
	}
	return s.UnmarshalText(data[1 : dataLength-1])
}

func (s State) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *State) UnmarshalText(b []byte) error {
	str := string(b)
	idx := strings.Index(_State_name, str)
	if idx < 0 {
		i, err := strconv.Atoi(str)
		if err == nil {
			*s = State(i)
			return nil
		}
		return fmt.Errorf("Cannot unmarshall \"%s\" to State. Is it mispelled?", str)
	}

	for i, v := range _State_index {
		if int(v) == idx {
			*s = State(i)
			return nil
		}
	}
	return fmt.Errorf("unexpected error in UnmarshalText for '%s' (go generate?)", s)
}

// ---- type ChargeState int

func (ch ChargeState) MarshalJSON() ([]byte, error) {
	b, err := ch.MarshalText()
	if err == nil {
		b = []byte(fmt.Sprintf("\"%s\"", string(b)))
	}
	return b, err
}

func (ch *ChargeState) UnmarshalJSON(data []byte) error {
	dataLength := len(data)
	if data[0] != '"' || data[dataLength-1] != '"' {
		return errors.New("ChargeState.UnmarshalJSON: Invalid JSON provided")
	}
	return ch.UnmarshalText(data[1 : dataLength-1])
}

func (ch ChargeState) MarshalText() ([]byte, error) {
	return []byte(ch.String()), nil
}

func (ch *ChargeState) UnmarshalText(b []byte) error {
	str := string(b)
	idx := strings.Index(_ChargeState_name, str)
	if idx < 0 {
		i, err := strconv.Atoi(str)
		if err == nil {
			*ch = ChargeState(i)
			return nil
		}
		return fmt.Errorf("Cannot unmarshall \"%s\" to ChargeState. Is it mispelled?", str)
	}

	for i, v := range _ChargeState_index {
		if int(v) == idx {
			*ch = ChargeState(i)
			return nil
		}
	}
	return fmt.Errorf("unexpected error in UnmarshalText for '%s' (go generate?)", ch)
}

// ---- type Mode int

func (m BotMode) MarshalJSON() ([]byte, error) {
	b, err := m.MarshalText()
	if err == nil {
		b = []byte(fmt.Sprintf("\"%s\"", string(b)))
	}
	return b, err
}

func (m *BotMode) UnmarshalJSON(data []byte) error {
	dataLength := len(data)
	if data[0] != '"' || data[dataLength-1] != '"' {
		return errors.New("State.UnmarshalJSON: Invalid JSON provided")
	}
	return m.UnmarshalText(data[1 : dataLength-1])
}

func (m BotMode) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

func (m *BotMode) UnmarshalText(b []byte) error {
	str := string(b)
	idx := strings.Index(_BotMode_name, str)
	if idx < 0 {
		i, err := strconv.Atoi(str)
		if err == nil {
			*m = BotMode(i)
			return nil
		}
		return fmt.Errorf("Cannot unmarshall \"%s\" to Mode. Is it mispelled?", str)
	}

	for i, v := range _BotMode_index {
		if int(v) == idx {
			*m = BotMode(i)
			return nil
		}
	}
	return fmt.Errorf("unexpected error in UnmarshalText for '%s' (go generate?)", m)
}
