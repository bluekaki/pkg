package rbt

import (
	// "fmt"
	"math/rand"
	_ "net/http/pprof"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())

	m.Run()
}

func TestXXX(t *testing.T) {
	for k := 0; k < 1; k++ {
		seeds := rand.Perm(100)

		tree := NewRbTree()
		for _, seed := range seeds {
			tree.Add(seed)
		}
		tree.Asc()

		for _, seed := range seeds {
			// t.Log(seed)
			// fmt.Println("-----------------", seed)
			// fmt.Println(tree.String())
			tree.Delete(seed, nil)
			tree.Asc()
		}
	}
}
