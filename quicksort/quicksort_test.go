package quicksort

import (
	"math/rand"
	"testing"
	"time"

	"github.com/bluekaki/pkg/stringutil"
)

type value int

func (x value) Compare(val Value) stringutil.Diff {
	y := val.(value)

	switch {
	case x < y:
		return stringutil.Less
	case x > y:
		return stringutil.Greater
	default:
		return stringutil.Equal
	}
}

func equal(x, y []Value) bool {
	for i := range x {
		if x[i].(value) != y[i].(value) {
			return false
		}
	}
	return true
}

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	m.Run()
}

func TestAsc(t *testing.T) {
	result := make([]Value, 2000)
	k := 0
	for i := 0; i < len(result); i += 2 {
		result[i] = value(k)
		result[i+1] = value(k)
		k++
	}

	for k := 0; k < 10000; k++ {
		seeds := rand.Perm(1000)

		values := make([]Value, 0, 2000)
		for k := 0; k < 2; k++ {
			for _, seed := range seeds {
				values = append(values, value(seed))
			}
		}

		Asc(values)
		if !equal(values, result) {
			t.Fatal("asc not match")
		}
	}
}

func TestDesc(t *testing.T) {
	result := make([]Value, 2000)
	k := 999
	for i := 0; i < len(result); i += 2 {
		result[i] = value(k)
		result[i+1] = value(k)
		k--
	}

	for k := 0; k < 10000; k++ {
		seeds := rand.Perm(1000)

		values := make([]Value, 0, 2000)
		for k := 0; k < 2; k++ {
			for _, seed := range seeds {
				values = append(values, value(seed))
			}
		}

		Desc(values)
		if !equal(values, result) {
			t.Fatal("desc not match")
		}
	}
}
