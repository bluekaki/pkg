package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bluekaki/pkg/rbt"
)

func main() {
	do := func() {
		rand.Seed(time.Now().UnixNano())
		seeds := rand.Perm(1000)

		tree := rbt.NewRbTree()
		check := func() {
			values := tree.Asc()
			for i := 1; i < len(values); i++ {
				if values[i] < values[i-1] {
					fmt.Println(values[:i+1])
					panic("not sorted in asc")
				}
			}
		}

		for _, seed := range seeds {
			tree.Add(seed)
		}
		check()
		if tree.Size() != uint32(len(seeds)) {
			fmt.Println(">>>>>>>>>", tree.Size())
			panic("not full")
		}

		for _, seed := range seeds {
			tree.Delete(seed)
			check()
		}
		if tree.Size() != 0 {
			fmt.Println(tree.String())
			panic("not empty")
		}
	}

	for k := 0; k < 10000; k++ {
		do()
		fmt.Println(">>", k)
	}
}
