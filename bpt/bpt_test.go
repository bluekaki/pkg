package bpt

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
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

func randSeed() int64 {
	buf := make([]byte, 8)
	io.ReadFull(crand.Reader, buf[:7])
	return int64(binary.BigEndian.Uint64(buf))
}

func TestMain(m *testing.M) {
	SetN(5)
	fmt.Println("_N:", _N, "_Mid:", _Mid, "_T:", _T)

	m.Run()
}

func mustContains(values []Value, target Value) {
	found := values[0].Compare(target) == stringutil.Equal
	for i := 1; i < len(values); i++ {
		if values[i].Compare(values[i-1]) != stringutil.Greater {
			panic("not in asc")
		}

		if values[i].Compare(target) == stringutil.Equal {
			found = true
		}
	}

	if !found {
		panic(fmt.Sprintf("%v not found", target))
	}
}

func mustNotContains(values []Value, target Value) {
	if len(values) == 0 {
		return
	}

	found := values[0].Compare(target) == stringutil.Equal
	for i := 1; i < len(values); i++ {
		if values[i].Compare(values[i-1]) != stringutil.Greater {
			panic("not in asc")
		}

		if values[i].Compare(target) == stringutil.Equal {
			found = true
		}
	}

	if found {
		panic(fmt.Sprintf("%v should not exist", target))
	}
}

func TestBPTInsert(t *testing.T) {
	for k := 0; k < 1000000; k++ {
		seed := randSeed()
		rand.Seed(seed)
		fmt.Println(">>>>", seed)

		values := rand.Perm(100)
		for i := range values[:50] {
			values[i] = -values[i]
		}

		tree := New()
		for _, v := range values {
			val := value(v)
			tree.Add(val)
			mustContains(tree.Asc(), val)
		}
		fmt.Println(k)
	}
}

func TestBPTDelete(t *testing.T) {
	for k := 0; k < 1000000; k++ {
		seed := randSeed()
		rand.Seed(seed)
		fmt.Println(">>>>", seed)

		values := rand.Perm(100)
		for i := range values[:50] {
			values[i] = -values[i]
		}

		tree := New()
		for _, v := range values {
			tree.Add(value(v))
		}
		// fmt.Println(tree)

		// tree.Delete(value(-15))
		// fmt.Println(tree)

		for _, v := range values {
			val := value(v)
			// fmt.Println("del", val)

			// fmt.Println("++++++++++++", tree)
			tree.Delete(val)
			// fmt.Println("============", tree)
			mustNotContains(tree.Asc(), val)
		}
		fmt.Println(k)
	}
}
