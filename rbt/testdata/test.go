package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/bluekaki/pkg/rbt"
)

var _ rbt.Value = (*Value)(nil)

type Value struct {
	val int
}

func (v *Value) ID() string {
	return strconv.Itoa(v.val)
}

func (v *Value) String() string {
	return strconv.Itoa(v.val)
}

func (v *Value) Compare(val rbt.Value) rbt.Diff {
	x := val.(*Value)
	switch {
	case v.val < x.val:
		return rbt.Less
	case v.val > x.val:
		return rbt.Greater
	default:
		return rbt.Equal
	}
}

func main() {
	do := func() {
		rand.Seed(time.Now().UnixNano())
		seeds := rand.Perm(1000)

		tree := rbt.NewRbTree()
		check := func() {
			values := tree.Asc()
			for i := 1; i < len(values); i++ {
				if values[i].Compare(values[i-1]) != rbt.Greater {
					fmt.Println(values[:i+1])
					panic("not sorted in asc")
				}
			}
		}

		for _, seed := range seeds {
			tree.Add(&Value{val: seed})
		}
		check()
		if tree.Size() != uint32(len(seeds)) {
			fmt.Println(">>>>>>>>>", tree.Size())
			panic("not full")
		}

		for _, seed := range seeds {
			tree.Delete(&Value{val: seed})
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
