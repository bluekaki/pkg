package rbt

import (
	crand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"testing"

	"github.com/bluekaki/pkg/stringutil"
)

var _ Value = (*value)(nil)

type value struct {
	Id  string
	Val int
}

func (v *value) ID() string {
	if v.Id == "" {
		return strconv.Itoa(int(v.Val))
	}

	return v.Id
}

func (v *value) String() string {
	return v.Id + "-" + strconv.Itoa(int(v.Val))
}

func (v *value) Compare(val Value) stringutil.Diff {
	x := val.(*value)
	switch {
	case v.Val < x.Val:
		return stringutil.Less
	case v.Val > x.Val:
		return stringutil.Greater
	default:
		return stringutil.Equal
	}
}

func (v *value) ToJSON() []byte {
	raw, _ := json.Marshal(v)
	return raw
}

func randSeed() int64 {
	buf := make([]byte, 8)
	io.ReadFull(crand.Reader, buf[:7])
	return int64(binary.BigEndian.Uint64(buf))
}

func mustContains(values []Value, target Value) {
	found := values[0].Compare(target) == stringutil.Equal
	for i := 1; i < len(values); i++ {
		if values[i].Compare(values[i-1]) == stringutil.Less {
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
		if values[i].Compare(values[i-1]) == stringutil.Less {
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

func TestSingleInsert(t *testing.T) {
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
			val := &value{Val: v}
			if tree.Exists(val) {
				t.Fatal("already exists")
			}
			if tree.ExistsByID(val) {
				t.Fatal("already exists")
			}

			if !tree.Add(val) {
				t.Fatal("insert nothing")
			}
			if tree.Add(val) {
				t.Fatal("duplicated")
			}

			if !tree.Exists(val) {
				t.Fatal("not found")
			}
			if !tree.ExistsByID(val) {
				t.Fatal("not found")
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

func TestMultiInsert(t *testing.T) {
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
			for index, prefix := range []string{"X", "Y"} {
				val := &value{Id: fmt.Sprintf("%s%d", prefix, v), Val: v}
				if index == 0 && tree.Exists(val) {
					t.Fatal("already exists")
				}
				if index == 1 && !tree.Exists(val) {
					t.Fatal("not exists")
				}

				if tree.ExistsByID(val) {
					t.Fatal("already exists")
				}

				if !tree.Add(val) {
					t.Fatal("insert nothing")
				}
				if tree.Add(val) {
					t.Fatal("duplicated")
				}

				if !tree.Exists(val) {
					t.Fatal("not found")
				}
				if !tree.ExistsByID(val) {
					t.Fatal("not found")
				}

				if tree.Size()-size != 1 {
					t.Fatal("size not match")
				}
				size = tree.Size()

				mustContains(tree.Asc(), val)
			}
		}
		fmt.Println(k)
	}
}

func TestSingleDelete(t *testing.T) {
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
			for _, prefix := range []string{"X", "Y"} {
				tree.Add(&value{Id: fmt.Sprintf("%s%d", prefix, v), Val: v})
			}
		}

		size := tree.Size()
		for _, v := range values {
			for index, prefix := range []string{"X", "Y"} {
				val := &value{Id: fmt.Sprintf("%s%d", prefix, v), Val: v}
				if !tree.Exists(val) {
					t.Fatal("not found")
				}
				if !tree.ExistsByID(val) {
					t.Fatal("not found")
				}

				if !tree.DeleteByID(val) {
					t.Fatal("delete nothing")
				}
				if tree.DeleteByID(val) {
					t.Fatal("duplicated")
				}

				if index == 0 && !tree.Exists(val) {
					t.Fatal("delete nothing")
				}
				if index == 1 && tree.Exists(val) {
					t.Fatal("delete nothing")
				}

				if tree.ExistsByID(val) {
					t.Fatal("delete nothing")
				}

				if size-tree.Size() != 1 {
					t.Fatal("size not match")
				}
				size = tree.Size()

				if index == 0 {
					mustContains(tree.Asc(), val)

				} else {
					mustNotContains(tree.Asc(), val)
				}
			}
		}
		fmt.Println(k)
	}
}

func TestMultiDelete(t *testing.T) {
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
			for _, prefix := range []string{"X", "Y"} {
				tree.Add(&value{Id: fmt.Sprintf("%s%d", prefix, v), Val: v})
			}
		}

		size := tree.Size()
		for _, v := range values {
			val := &value{Val: v}
			if !tree.Exists(val) {
				t.Fatal("not found")
			}

			if !tree.Delete(val) {
				t.Fatal("delete nothing")
			}
			if tree.Delete(val) {
				t.Fatal("duplicated")
			}

			if tree.Exists(val) {
				t.Fatal("delete nothing")
			}
			if tree.ExistsByID(val) {
				t.Fatal("delete nothing")
			}

			if size-tree.Size() != 2 {
				t.Fatal("size not match")
			}
			size = tree.Size()

			mustNotContains(tree.Asc(), val)
		}
		fmt.Println(k)
	}
}

func TestToJSON(t *testing.T) {
	seed := randSeed()
	rand.Seed(seed)
	fmt.Println(">>>>", seed)

	values := rand.Perm(10)
	for i := range values[:5] {
		values[i] = -values[i]
	}

	tree := New()
	for _, v := range values {
		tree.Add(&value{Val: v})
	}
	fmt.Println("00", tree)
	fmt.Println(string(tree.ToJSON()))

	rbt, err := JSON2Tree(tree.ToJSON(), func(raw []byte) (Value, error) {
		val := new(value)
		err := json.Unmarshal(raw, val)
		return val, err
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("11", rbt)
}
