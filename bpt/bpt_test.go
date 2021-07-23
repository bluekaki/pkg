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

var mp = make(map[int]int)

func TestMain(m *testing.M) {
	SetN(5)
	fmt.Println("_N:", _N, "_Mid:", _Mid, "_T:", _T)

	seed := randSeed()
	rand.Seed(seed)
	fmt.Println(">>seed<<", seed)

	index := 1
	for k := 'A'; k <= 'Z'; k++ {
		mp[int(k)] = index
		index++
	}

	m.Run()
}

func TestBPTInsert(t *testing.T) {
	if false {
		tree := New()

		values := rand.Perm(40)
		for i := range values[:20] {
			values[i] = -values[i]
		}

		for _, v := range values {
			tree.Add(value(v))
		}
		fmt.Println(tree)
	}

	if false {
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
}

func TestBPTDelete(t *testing.T) {
	// tree := New()
	// tree.Add(value(mp['H']))
	// tree.Add(value(mp['P']))
	// tree.Add(value(mp['C']))
	// tree.Add(value(mp['G']))
	// tree.Add(value(mp['M']))
	// tree.Add(value(mp['T']))
	// tree.Add(value(mp['X']))
	// tree.Add(value(mp['A']))
	// tree.Add(value(mp['B']))
	// tree.Add(value(mp['D']))
	// tree.Add(value(mp['E']))
	// tree.Add(value(mp['F']))
	// tree.Add(value(mp['J']))
	// tree.Add(value(mp['K']))
	// tree.Add(value(mp['L']))
	// tree.Add(value(mp['N']))
	// tree.Add(value(mp['O']))
	// tree.Add(value(mp['Q']))
	// tree.Add(value(mp['R']))
	// tree.Add(value(mp['S']))
	// tree.Add(value(mp['U']))
	// tree.Add(value(mp['V']))
	// tree.Add(value(mp['Y']))
	// tree.Add(value(mp['Z']))
	// fmt.Println(tree)

	// rand.Seed(1626955704130035742)

	for k := 0; k < 1000000; k++ {
		seed := randSeed()
		fmt.Println(">>>>", seed)
		// rand.Seed(-2084703646357133056)

		values := rand.Perm(100)
		for i := range values[:50] {
			values[i] = -values[i]
		}

		tree := New()
		for _, v := range values {
			tree.Add(value(v))
		}

		fmt.Println(k)
		tree.Asc()

		for _, v := range values {
			// fmt.Println("del:", v)
			tree.Delete(value(v))
			tree.Asc()
		}

	}
}
