package regenbox

import "bytes"

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
