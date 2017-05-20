package regenbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func expect(t *testing.T, test, v, to string) {
	if v != to {
		t.Errorf("%s: expected \"%s\" to equal \"%s\".", test, v, to)
	}
}

func TestTypesMarshallers(t *testing.T) {
	var (
		s        State
		ch       ChargeState
		m        BotMode
		expected string
		b        []byte
		err      error
	)

	s = State(Connected)
	expected = fmt.Sprintf("\"%s\"", s)
	b, err = json.Marshal(s)
	if err != nil {
		t.Error(err)
	} else {
		expect(t, "State_MarshallJSON", string(b), string(expected))
	}

	ch = ChargeState(Charging)
	expected = fmt.Sprintf("\"%s\"", ch)
	b, err = json.Marshal(ch)
	if err != nil {
		t.Error(err)
	} else {
		expect(t, "ChargeState_MarshallJSON", string(b), string(expected))
	}

	m = BotMode(Charger)
	expected = fmt.Sprintf("\"%s\"", m)
	b, err = json.Marshal(m)
	if err != nil {
		t.Error(err)
	}
	expect(t, "Mode_MarshallJSON", string(b), string(expected))
}

func TestUnmarshallers(t *testing.T) {
	var (
		s   State
		ch  ChargeState
		m   BotMode
		b   *bytes.Buffer
		dec *json.Decoder
		err error
	)

	b = new(bytes.Buffer)
	b.WriteString("\"Connected\"")
	dec = json.NewDecoder(b)
	err = dec.Decode(&s)
	if err != nil {
		t.Error(err)
	} else {
		expect(t, "State_UnmarshallJSON", s.String(), Connected.String())
	}

	b = new(bytes.Buffer)
	b.WriteString("\"Charging\"")
	dec = json.NewDecoder(b)
	err = dec.Decode(&ch)
	if err != nil {
		t.Error(err)
	} else {
		expect(t, "ChargeState_UnmarshallJSON", ch.String(), Charging.String())
	}

	b = new(bytes.Buffer)
	b.WriteString("\"Charger\"")
	dec = json.NewDecoder(b)
	err = dec.Decode(&m)
	if err != nil {
		t.Error(err)
	} else {
		expect(t, "ChargeState_UnmarshallJSON", m.String(), Charger.String())
	}
}
