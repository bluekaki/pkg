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
	seeds := rand.Perm(1000000)

	tree := NewRbTree()
	for _, seed := range seeds {
		tree.Add(seed)
	}
	tree.Asc()
}
