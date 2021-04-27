package crypto

import (
	"crypto/rand"
	"io"

	"github.com/byepichi/pkg/minami58"
)

// A32RandomID  a 32 bytes random id
func A32RandomID() string {
	buf := make([]byte, 16)
	io.ReadFull(rand.Reader, buf)

	return string(minami58.Encode(buf))
}
