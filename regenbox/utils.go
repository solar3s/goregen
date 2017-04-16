package regenbox

import (
	"bytes"
	"encoding/binary"
)

var crlf = []byte{13, 10}

func trimCRLF(buf []byte) []byte {
	i := len(buf)
	if i < 2 {
		return buf
	}
	if 0 == bytes.Compare(buf[i-2:i], crlf) {
		return buf[:i-2]
	}
	return buf
}

type LedState bool

func (led LedState) String() string {
	if led {
		return "on"
	}
	return "off"
}

func readUint(b []byte) (uint64, error) {
	buf := bytes.NewBuffer(b)
	if ui64, err := binary.ReadUvarint(buf); err != nil {
		return 0, err
	} else {
		return ui64, nil
	}
}
