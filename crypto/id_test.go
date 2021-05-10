package crypto

import (
	"testing"
)

func TestIDGenerator(t *testing.T) {
	var prefix IDPrefix = [2]byte{'A', 'X'}
	for k := 0; k < 100; k++ {
		id := A20RID(prefix)
		t.Log(id)
	}
}
