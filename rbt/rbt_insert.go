package rbt

func (t *rbTree) Add(value int) {
	if t.root == nil {
		t.root = &node{
			color: black,
			val:   value,
		}
		return
	}

	x := t.add(value)

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
		// fmt.Println("left left left left left left left left left left")

		g := x.G()
		g.color = red

		x.P.color = black
		t3 := x.P.R
		if t3 != nil {
			t3.P = g
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
		g.L = t3

	case x.P.Left() && x.Right():
		// fmt.Println("left right left right left right left right left right")

		p := x.P
		t1 := x.L
		if t1 != nil {
			t1.P = p
		}
		p.R = t1

		g := x.G()
		g.L = x
		x.P = g

		x.L = p
		p.P = x

		x = p

		goto sw

	case x.P.Right() && x.Right():
		// fmt.Println("right right right right right right right right right right")

		g := x.G()
		g.color = red

		x.P.color = black
		t3 := x.P.L
		if t3 != nil {
			t3.P = g
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
		g.R = t3

	case x.P.Right() && x.Left():
		// fmt.Println("right left right left right left right left right left right left ")

		p := x.P
		t4 := x.R
		if t4 != nil {
			t4.P = p
		}
		p.L = t4

		g := x.G()
		g.R = x
		x.P = g

		x.R = p
		p.P = x

		x = p

		goto sw
	}
}

func (t *rbTree) add(value int) *node {
	root := t.root
	for {
		if value < root.val {
			if root.L != nil {
				root = root.L

			} else {
				root.L = &node{
					P:     root,
					color: red,
					val:   value,
				}
				return root.L
			}

		} else if value > root.val {
			if root.R != nil {
				root = root.R

			} else {
				root.R = &node{
					P:     root,
					color: red,
					val:   value,
				}
				return root.R
			}

		} else {
			return nil
		}
	}
}
