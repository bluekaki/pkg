package rbt

import (
	"github.com/bluekaki/pkg/stringutil"
)

func (t *rbTree) Add(val Value) {
	if val == nil {
		return
	}

	t.Lock()
	defer t.Unlock()

	if t.root == nil {
		t.root = &node{
			color:  black,
			values: []Value{val},
		}
		t.size++
		return
	}

	x := t.add(val)
	if x == nil {
		// duplicated
		return
	}
	t.size++

loop:
	if x.Root() {
		x.color = black
		return
	}

	if x.P.Balck() {
		// do nothing
		return
	}

	u := x.U()
	if u == nil {
		return
	}

	if u.Red() {
		x.P.color = black
		u.color = black

		g := x.G()
		g.color = red

		x = g
		goto loop
	}

sw:
	switch {
	case x.P.Left() && x.Left():
		g := x.G()
		g.color = red

		x.P.color = black
		y := x.P.R
		if y != nil {
			y.P = g
		}

		x.P.P = g.P
		if g.P == nil {
			t.root = x.P

		} else {
			if g.Left() {
				g.P.L = x.P
			} else {
				g.P.R = x.P
			}
		}

		x.P.R = g
		g.P = x.P
		g.L = y

	case x.P.Left() && x.Right():
		p := x.P
		y := x.L
		if y != nil {
			y.P = p
		}
		p.R = y

		g := x.G()
		g.L = x
		x.P = g

		x.L = p
		p.P = x

		x = p

		goto sw

	case x.P.Right() && x.Right():
		g := x.G()
		g.color = red

		x.P.color = black
		y := x.P.L
		if y != nil {
			y.P = g
		}

		x.P.P = g.P
		if g.P == nil {
			t.root = x.P

		} else {
			if g.Left() {
				g.P.L = x.P
			} else {
				g.P.R = x.P
			}
		}

		x.P.L = g
		g.P = x.P
		g.R = y

	case x.P.Right() && x.Left():
		p := x.P
		y := x.R
		if y != nil {
			y.P = p
		}
		p.L = y

		g := x.G()
		g.R = x
		x.P = g

		x.R = p
		p.P = x

		x = p

		goto sw
	}
}

func (t *rbTree) add(val Value) *node {
	root := t.root
	for {
		switch val.Compare(root.values[0]) {
		case stringutil.Less:
			if root.L != nil {
				root = root.L

			} else {
				root.L = &node{
					P:      root,
					color:  red,
					values: []Value{val},
				}
				return root.L
			}

		case stringutil.Greater:
			if root.R != nil {
				root = root.R

			} else {
				root.R = &node{
					P:      root,
					color:  red,
					values: []Value{val},
				}
				return root.R
			}

		case stringutil.Equal:
			for _, value := range root.values {
				if val.ID() == value.ID() {
					// duplicated
					return nil
				}
			}

			root.values = append(root.values, val)
			return root
		}
	}
}
