package rbt

import (
	"bytes"
	"container/list"
	"sync"
)

func NewRbTree() *rbTree {
	return new(rbTree)
}

type rbTree struct {
	sync.RWMutex
	size uint32
	root *node
}

func (t *rbTree) String() string {
	t.RLock()
	defer t.RUnlock()

	buf := bytes.NewBufferString("RedBlackTree\n")
	if t.root != nil {
		output(t.root, "", true, buf)
	}
	return buf.String()
}

func output(root *node, prefix string, isTail bool, buf *bytes.Buffer) {
	if root.R != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "│   "
		} else {
			newPrefix += "    "
		}

		output(root.R, newPrefix, false, buf)
	}

	buf.WriteString(prefix)
	if isTail {
		buf.WriteString("└── ")
	} else {
		buf.WriteString("┌── ")
	}

	buf.WriteString(root.String())
	buf.WriteString("\n")

	if root.L != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}

		output(root.L, newPrefix, true, buf)
	}
}

func (t *rbTree) Empty() bool {
	t.RLock()
	defer t.RUnlock()

	return t.size == 0
}

func (t *rbTree) Size() uint32 {
	t.RLock()
	defer t.RUnlock()

	return t.size
}

func (t *rbTree) Asc() []int {
	t.RLock()
	defer t.RUnlock()

	values := make([]int, 0, t.size)

	stack := list.New()
	root := t.root
	for root != nil || stack.Len() != 0 {
		for root != nil {
			stack.PushBack(root)
			root = root.L
		}

		if stack.Len() != 0 {
			v := stack.Back()
			root = v.Value.(*node)
			values = append(values, root.val) //visit
			root = root.R
			stack.Remove(v)
		}
	}

	return values
}
