package crypto

import (
	"crypto/rand"
	"encoding/binary"
	"io"
	"strconv"
)

// IDPrefix ...
type IDPrefix [2]byte

func (i IDPrefix) String() string {
	return string(i[:])
}

// A20RID  a 18len decimal random id with 2len prefix
func A20RID(prefix IDPrefix) string {
	id := prefix.String()
	buf := make([]byte, 8)
	for len(id) <= 20 {
		io.ReadFull(rand.Reader, buf)
		id += strconv.FormatUint(binary.BigEndian.Uint64(buf), 10)
	}

	return id[:20]
}
