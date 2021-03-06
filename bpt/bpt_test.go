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
	"github.com/bluekaki/pkg/zaplog"
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

func (v value) ToJSON() []byte {
	return []byte(strconv.Itoa(int(v)))
}

func randSeed() int64 {
	buf := make([]byte, 8)
	io.ReadFull(crand.Reader, buf[:7])
	return int64(binary.BigEndian.Uint64(buf))
}

var logger, _ = zaplog.NewJSONLogger()

func TestMain(m *testing.M) {
	defer logger.Sync()

	m.Run()
}

func NewTree(t uint16) BPTree {
	return New(t, "/data/bpt", logger, func(raw []byte) Value {
		v, _ := strconv.ParseInt(string(raw), 10, 64)
		return value(v)
	})
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

		tree := NewTree(10)
		size := tree.Size()
		for _, v := range values {
			val := value(v)
			if !tree.Add(val) {
				t.Fatal("insert nothing")
			}

			if tree.Add(val) {
				t.Fatal("duplicated")
			}

			if tree.Size()-size != 1 {
				t.Fatal("size not match")
			}
			size = tree.Size()

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

		tree := NewTree(10)
		for _, v := range values {
			tree.Add(value(v))
		}

		rand.Shuffle(len(values), func(i, j int) {
			values[i], values[j] = values[j], values[i]
		})

		size := tree.Size()
		for _, v := range values {
			val := value(v)
			if !tree.Delete(val) {
				t.Fatal("delete nothing")
			}

			if tree.Delete(val) {
				t.Fatal("duplicated")
			}

			if size-tree.Size() != 1 {
				t.Fatal("size not match")
			}
			size = tree.Size()

			mustNotContains(tree.Asc(), val)
		}
		fmt.Println(k)
	}
}
