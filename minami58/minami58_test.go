package minami58

import (
	"bytes"
	"crypto/rand"
	"io"
	mrand "math/rand"
	"strconv"
	"testing"
	"time"
)

func TestEncode(t *testing.T) {
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

func BenchmarkEncode(b *testing.B) {
	payload := make([]byte, 1027)
	io.ReadFull(rand.Reader, payload)

	raw := Decode(Encode(payload))
	if !bytes.Equal(payload, raw) {
		b.Fatal("not match")
	}
}

func TestShorten(t *testing.T) {
	total := 30000000
	index := make(map[string]struct{}, total)

	duplicated := 0
	for k := 0; k < total; k++ {
		link := Shorten(strconv.Itoa(k))
		if _, ok := index[link]; ok {
			duplicated++

			link = Shorten(strconv.Itoa(k) + " ")
			if _, ok = index[link]; ok {
				t.Fatal(strconv.Itoa(k), link)
			}
		}
		index[link] = struct{}{}
	}

	t.Log("duplicated", duplicated)
}

func BenchmarkShorten(b *testing.B) {
	payload := make([]byte, 1027)
	io.ReadFull(rand.Reader, payload)

	Shorten(string(payload))
}
