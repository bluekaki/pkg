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
	// t.RLock()
	// defer t.RUnlock()

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

func (t *bpTree) Asc() (values []Value) {
	t.RLock()
	defer t.RUnlock()

	if t.root == nil {
		return
	}

	type item struct {
		*node
		cur int
	}

	stack := list.New()
	stack.PushBack(&item{node: t.root})

	for stack.Len() > 0 {
		element := stack.Back()
		stack.Remove(element)

		node := element.Value.(*item)
		if node.node != t.root && len(node.values) < (_T-1) {
			fmt.Println(">>", node.id, t.String())
			panic("-----illegal------")
		}

		if node.leaf() {
			values = append(values, node.values...)
			continue
		}

		if node.cur <= len(node.values) {
			child := &item{node: node.children[node.cur]}

			if node.cur > 0 {
				values = append(values, node.values[node.cur-1])
			}
			node.cur++
			stack.PushBack(node)
			stack.PushBack(child)
		}
	}
	return
}
