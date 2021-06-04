package rbt

import (
	// "fmt"
	"math/rand"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())

	m.Run()
}

func TestXXX(t *testing.T) {
	for k := 0; k < 10000; k++ {
		seeds := rand.Perm(200)

		tree := NewRbTree()
		for _, seed := range seeds {
			tree.Add(seed)
		}
		tree.Asc()

		for _, seed := range seeds {
			// t.Log(seed)
			// fmt.Println("-----------------", seed)
			// fmt.Println(tree.String())
			tree.Delete(seed)
			tree.Asc()
		}
	}
}
