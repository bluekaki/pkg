package minami58

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

func Test(t *testing.T) {
	payload := make([]byte, 256)

	for k := 0; k < 1024; k++ {
		io.ReadFull(rand.Reader, payload)

		raw := Encode(payload)
		t.Log(string(raw))

		raw, err := Decode(raw)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(payload, raw) {
			t.Fatal("not match")
		}
	}
}

func Benchmark(b *testing.B) {
	payload := make([]byte, 256)
	io.ReadFull(rand.Reader, payload)

	raw, err := Decode(Encode(payload))
	if err != nil {
		b.Fatal(err)
	}

	if !bytes.Equal(payload, raw) {
		b.Fatal("not match")
	}
}
