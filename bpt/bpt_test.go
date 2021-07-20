package bpt

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/bluekaki/pkg/stringutil"
)

type value int

func (v value) String() string {
	return strconv.Itoa(int(v))
}

func (x value) Compare(v Value) stringutil.Diff {
	y := v.(value)

	if x < y {
		return stringutil.Less
	} else if x > y {
		return stringutil.Greater
	} else {
		return stringutil.Equal
	}
}

func TestMain(m *testing.M) {
	SetN(5)

	fmt.Println("_N:", _N, "_Mid:", _Mid, "_T:", _T)
	m.Run()
}

func TestBPT(t *testing.T) {
	tree := New()

	for k := 0; k < 20; k++ {
		tree.Add(value(k))
		fmt.Println(tree)
	}

	for k := -20; k <= 0; k++ {
		tree.Add(value(k))
		fmt.Println(tree)
	}
}
