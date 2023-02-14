package id

import (
	"testing"
)

type SEQ struct {
	val uint32
}

func (s *SEQ) Next() (uint32, error) {
	s.val++
	return s.val, nil
}

var gen Generator
var prefix Prefix = [2]byte{'A', 'X'}

func TestMain(m *testing.M) {
	gen = NewGenerator(new(SEQ))
	m.Run()
}

func TestGenrator(t *testing.T) {
	for k := 0; k < 100; k++ {
		id, err := gen.New(prefix)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(id)

		prefix, _, err := gen.Parse(id)
		if err != nil {
			t.Fatal(err)
		}
		if prefix.String() != "AX" {
			t.Fatal("prefix not match")
		}
	}
}

func BenchmarkGenrator(b *testing.B) {
	id, err := gen.New(prefix)
	if err != nil {
		b.Fatal(err)
	}

	prefix, _, err := gen.Parse(id)
	if err != nil {
		b.Fatal(err)
	}
	if prefix.String() != "AX" {
		b.Fatal("prefix not match")
	}
}
