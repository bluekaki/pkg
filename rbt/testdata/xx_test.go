package testdata

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/bluekaki/pkg/rbt"
	"github.com/bluekaki/pkg/stringutil"
)

var _ rbt.Value = (*Value)(nil)

type Value struct {
	val uint16
}

func (v *Value) ID() string {
	return strconv.Itoa(int(v.val))
}

func (v *Value) String() string {
	return strconv.Itoa(int(v.val))
}

func (v *Value) Compare(val rbt.Value) stringutil.Diff {
	x := val.(*Value)
	switch {
	case v.val < x.val:
		return stringutil.Less
	case v.val > x.val:
		return stringutil.Greater
	default:
		return stringutil.Equal
	}
}

func (v *Value) Marshal() []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, v.val)

	return buf
}

func Unmarshal(raw []byte) rbt.Value {
	return &Value{
		val: binary.BigEndian.Uint16(raw),
	}
}

func TestXXX(t *testing.T) {
	do := func() {
		rand.Seed(time.Now().UnixNano())
		seeds := rand.Perm(1000)

		tree := rbt.New()
		check := func() {
			values := tree.Asc()
			for i := 1; i < len(values); i++ {
				if values[i].Compare(values[i-1]) != stringutil.Greater {
					fmt.Println(values[:i+1])
					t.Fatal("not sorted in asc")
				}
			}
		}

		for _, seed := range seeds {
			tree.Add(&Value{val: uint16(seed)})
		}
		check()
		if tree.Size() != uint32(len(seeds)) {
			fmt.Println(">>>>>>>>>", tree.Size())
			t.Fatal("not full")
		}

		for _, seed := range seeds {
			tree.Delete(&Value{val: uint16(seed)})
			check()
		}
		if tree.Size() != 0 {
			fmt.Println(tree.String())
			t.Fatal("not empty")
		}
	}

	for k := 0; k < 10000; k++ {
		do()
		fmt.Println(">>", k)
	}
}

func TestSort(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	seeds := rand.Perm(20)

	tree := rbt.New()
	for _, seed := range seeds {
		tree.Add(&Value{val: uint16(seed)})
	}

	fmt.Println(tree.Asc())
	fmt.Println(tree.Desc())
	fmt.Println(tree.String())
}

func TestMarshal(t *testing.T) {
	do := func() {
		rand.Seed(time.Now().UnixNano())
		seeds := rand.Perm(1000)

		tree := rbt.New()
		check := func() {
			values := tree.Asc()
			for i := 1; i < len(values); i++ {
				if values[i].Compare(values[i-1]) != stringutil.Greater {
					fmt.Println(values[:i+1])
					t.Fatal("not sorted in asc")
				}
			}
		}

		for _, seed := range seeds {
			tree.Add(&Value{val: uint16(seed)})
		}

		raw := tree.Marshal()
		tree, err := rbt.Unmarshal(raw, Unmarshal)
		if err != nil {
			t.Fatal(err)
		}

		check()
	}

	for k := 0; k < 10000; k++ {
		do()
		fmt.Println(">>", k)
	}
}
