package rbt

import (
	"fmt"
)

type color uint8

const (
	red   color = 1
	black color = 2
)

type node struct {
	color   color
	val     int
	L, R, P *node
}

func (n *node) Red() bool {
	return n.color == red
}

func (n *node) Balck() bool {
	return n.color == black
}

func (n *node) Root() bool {
	return n.P == nil
}

func (n *node) Left() bool {
	return n.P.L == n
}

func (n *node) Right() bool {
	return n.P.R == n
}

func (n *node) G() *node {
	if n.P == nil {
		return nil
	}
	return n.P.P
}

func (n *node) U() *node {
	g := n.G()
	if g == nil {
		return nil
	}

	if g.L == n.P {
		return g.R
	}
	return g.L
}

func (n *node) S() *node {
	p := n.P
	if p == nil {
		return nil
	}

	if p.L == n {
		return p.R
	}
	return p.L
}

func NewRbTree() *rbTree {
	return new(rbTree)
}

type rbTree struct {
	root *node
}

func (t *rbTree) Asc() {
	var values []int
	push := func(value int) {
		values = append(values, value)
	}

	asc(t.root, push)

	for i := 1; i < len(values); i++ {
		if values[i]-values[i-1] != 1 {
			fmt.Println(values[:i+1])
			panic("xxx")
		}
	}
}

func asc(root *node, push func(value int)) {
	if root != nil {
		asc(root.L, push)
		push(root.val)
		asc(root.R, push)
	}
}
