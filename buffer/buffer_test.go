package buffer

import (
	"testing"
)

func TestFixedBuffer(t *testing.T) {
	buf := NewFixedBuffer(10)
	t.Log(buf.Bytes())

	for k := byte(0); k < 15; k++ {
		buf.Write([]byte{k})
		t.Log(buf.Bytes())
	}
}
