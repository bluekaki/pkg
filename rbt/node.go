package rbt

import (
	"strconv"
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

func (n *node) String() string {
	if n.Red() {
		return strconv.Itoa(n.val) + "r"

	} else {
		return strconv.Itoa(n.val) + "b"
	}
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
