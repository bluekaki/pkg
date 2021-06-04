package rbt

func (t *rbTree) Delete(value int) {
	x := t.lookup(value)
	if x == nil {
		return
	}

	// u := x

del:
	switch {
	case x.L == nil && x.R == nil:
		if x.Root() {
			t.root = nil
			return
		}

		// u.val = x.val
		if x.Left() {
			x.P.L = nil

		} else {
			x.P.R = nil
		}

	case x.L == nil || x.R == nil:
		if x.L == nil {
			s := t.minimum(x.R)
			x.val = s.val
			x = s

		} else {
			s := t.maximum(x.L)
			x.val = s.val
			x = s
		}
		goto del

	default:
		s := t.minimum(x.R)
		x.val = s.val
		x = s
		goto del
	}

	// if x.Red() || x.P.Red() {
	// 	x.P.color = black
	// 	return
	// }
}

func (t *rbTree) lookup(value int) *node {
	root := t.root
	for root != nil {
		if value < root.val {
			root = root.L

		} else if value > root.val {
			root = root.R

		} else {
			return root
		}
	}

	return nil
}

func (t *rbTree) maximum(root *node) *node {
	for root.R != nil {
		root = root.R
	}
	return root
}

func (t *rbTree) minimum(root *node) *node {
	for root.L != nil {
		root = root.L
	}
	return root
}
