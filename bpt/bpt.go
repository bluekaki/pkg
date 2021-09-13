package bpt

import (
	"bytes"
	"container/list"
	"fmt"
	"strings"
	"sync"
)

func New(orderT uint16) *bpTree {
	if orderT%2 != 0 {
		panic("t must be even number")
	}

	if orderT < 4 { // t ≥4
		panic("t must be ≥4")
	}

	bpt := new(bpTree)
	bpt.meta.N = int(orderT - 1)
	bpt.meta.Mid = int(orderT-1) / 2
	bpt.meta.HT = int(orderT) / 2

	return bpt
}

type bpTree struct {
	sync.RWMutex
	size uint32
	root *node

	meta struct {
		N   int // the max values in one node
		Mid int // N / 2
		HT  int // (N + 1) / 2  (half order)the half children in one node
	}
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
			stack.PushBack(fmt.Sprintf("%s%s\n", strings.Repeat("    ", level), node.values[e].String()))
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
		if node.node != t.root && len(node.values) < (t.meta.HT-1) {
			panic(fmt.Sprintf("illegal %s", t.String()))
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
