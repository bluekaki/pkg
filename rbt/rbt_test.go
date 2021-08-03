package rbt

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

var _ Value = (*value)(nil)

type value struct {
	val int
}

func (v *value) ID() string {
	return strconv.Itoa(int(v.val))
}

func (v *value) String() string {
	return strconv.Itoa(int(v.val))
}

func (v *value) Compare(val Value) stringutil.Diff {
	x := val.(*value)
	switch {
	case v.val < x.val:
		return stringutil.Less
	case v.val > x.val:
		return stringutil.Greater
	default:
		return stringutil.Equal
	}
}

func randSeed() int64 {
	buf := make([]byte, 8)
	io.ReadFull(crand.Reader, buf[:7])
	return int64(binary.BigEndian.Uint64(buf))
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

func TestInsert(t *testing.T) {
	for k := 0; k < 1000000; k++ {
		seed := randSeed()
		rand.Seed(seed)
		fmt.Println(">>>>", seed)

		values := rand.Perm(1000)
		for i := range values[:500] {
			values[i] = -values[i]
		}

		tree := New()
		size := tree.Size()
		for _, v := range values {
			val := &value{val: v}
			if !tree.Add(val) {
				t.Fatal("insert nothing")
			}

			if tree.Add(val) {
				t.Fatal("duplicated")
			}

			if tree.Size()-size != 1 {
				t.Fatal("size not match")
			}
			size = tree.size

			mustContains(tree.Asc(), val)
		}
		fmt.Println(k)
	}
}

func TestDelete(t *testing.T) {
	for k := 0; k < 1000000; k++ {
		seed := randSeed()
		rand.Seed(seed)
		fmt.Println(">>>>", seed)

		values := rand.Perm(1000)
		for i := range values[:500] {
			values[i] = -values[i]
		}

		tree := New()
		for _, v := range values {
			tree.Add(&value{val: v})
		}

		size := tree.Size()
		for _, v := range values {
			val := &value{val: v}
			if !tree.Delete(val) {
				t.Fatal("delete nothing")
			}

			if tree.Delete(val) {
				t.Fatal("duplicated")
			}

			if size-tree.Size() != 1 {
				t.Fatal("size not match")
			}
			size = tree.size

			mustNotContains(tree.Asc(), val)
		}
		fmt.Println(k)
	}
}
