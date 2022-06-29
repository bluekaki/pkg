package id

import (
	"crypto/rand"
	"io"

	"github.com/bluekaki/pkg/minami58"
)

// JournalID  a minami58 encoded random string
func JournalID() string {
	nonce := make([]byte, 30)
	io.ReadFull(rand.Reader, nonce)

	return string(minami58.Encode(nonce))
}
