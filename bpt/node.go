package bpt

import (
	"github.com/bluekaki/pkg/stringutil"
)

type Value interface {
	String() string
	Compare(Value) stringutil.Diff
	// Marshal() []byte
}

type node struct {
	id       int
	values   []Value
	children []*node
}

func (n *node) full() bool {
	return len(n.values) == _N
}

func (n *node) overHalf() bool {
	return len(n.values) >= _T
}

func (n *node) leaf() bool {
	return len(n.children) == 0
}
