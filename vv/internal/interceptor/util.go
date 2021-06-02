package interceptor

import (
	"crypto/rand"
	"io"
	"net/url"

	"github.com/bluekaki/pkg/minami58"
)

func QueryUnescape(uri string) string {
	decodedUri, err := url.QueryUnescape(uri)
	if err != nil {
		return uri
	}

	return decodedUri
}

func GenJournalID() string {
	nonce := make([]byte, 16)
	io.ReadFull(rand.Reader, nonce)

	return string(minami58.Encode(nonce))
}
