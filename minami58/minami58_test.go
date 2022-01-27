package minami58

import (
	"bytes"
	"crypto/rand"
	"io"
	mrand "math/rand"
	"testing"
	"time"
)

func Test(t *testing.T) {
	rander := mrand.New(mrand.NewSource(time.Now().UnixNano()))

	for k := 0; k < 1000000; k++ {
		buf := make([]byte, rander.Intn(100))
		io.ReadFull(rand.Reader, buf)

		raw := Decode((Encode(buf)))
		if !bytes.Equal(buf, raw) {
			t.Log(buf)
			t.Log(raw)

			t.Fatal("not match")
		}
	}
}

func Benchmark(b *testing.B) {
	payload := make([]byte, 1027)
	io.ReadFull(rand.Reader, payload)

	raw := Decode(Encode(payload))
	if !bytes.Equal(payload, raw) {
		b.Fatal("not match")
	}
}
