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

func (n *node) full() bool {
	return len(n.values) == _N
}
