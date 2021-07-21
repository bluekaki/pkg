package bpt

import (
	"bytes"
	"container/list"
	"fmt"
	"strings"
	"sync"
)

var (
	_N   = 199 // the max values in one node
	_Mid = _N / 2
	_T   = (_N + 1) / 2 // the half children in one node
	// _Limit = _T - 1

	once sync.Once
)

// SetN N should be odd number
func SetN(n uint16) {
	once.Do(func() {
		if n%2 == 0 {
			panic("n must be odd number")
		}

		if n < 3 { // t ≥2; 2t-1 = n;
			panic("n must be ≥3")
		}

		_N = int(n)
		_Mid = _N / 2
		_T = (_N + 1) / 2
		// _Limit = _T - 1
	})
}

func New() *bpTree {
	return new(bpTree)
}

type bpTree struct {
	sync.RWMutex
	size uint32
	root *node
}

func (t *bpTree) String() string {
	t.RLock()
	defer t.RUnlock()

	stack := list.New()
	if t.root != nil {
		output(stack, t.root, 0, true)
	}

	buf := bytes.NewBufferString(fmt.Sprintf("BTree %d\n", t.size))
	for stack.Len() > 0 {
		element := stack.Back()
		stack.Remove(element)

		buf.WriteString(element.Value.(string))
	}

	return buf.String()
}

func output(stack *list.List, node *node, level int, isTail bool) {
	for e := 0; e < len(node.values)+1; e++ {
		if e < len(node.children) {
			output(stack, node.children[e], level+1, true)
		}

		if e < len(node.values) {
			stack.PushBack(fmt.Sprintf("%s%s(%d)\n", strings.Repeat("    ", level), node.values[e].String(), node.id))
		}
	}
}

func (t *bpTree) Empty() bool {
	t.RLock()
	defer t.RUnlock()

	return t.size == 0
}

func (t *bpTree) Size() uint32 {
	t.RLock()
	defer t.RUnlock()

	return t.size
}

var id int

func ID() int {
	id++
	return id
}
