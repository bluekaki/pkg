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
	values   []Value
	children []*node
}

func (n *node) full(N int) bool {
	return len(n.values) == N
}

func (n *node) overHalf(HT int) bool {
	return len(n.values) >= HT
}

func (n *node) leaf() bool {
	return len(n.children) == 0
}
