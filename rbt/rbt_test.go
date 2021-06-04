package rbt

import (
	"math/rand"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())

	m.Run()
}

func TestXXX(t *testing.T) {
	for k := 0; k < 1000000; k++ {
		seeds := rand.Perm(100)

		tree := NewRbTree()
		for _, seed := range seeds {
			tree.Add(seed)
		}
		tree.Asc()

		for _, seed := range seeds {
			tree.Delete(seed)
			tree.Asc()
		}
	}
}
