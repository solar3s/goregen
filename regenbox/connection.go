package regenbox

import "io"

type Connection interface {
	io.ReadWriteCloser
}
