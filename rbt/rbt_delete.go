package rbt

func (t *rbTree) Delete(value int) {
	x := t.lookup(value)
	if x == nil {
		return
	}

del:
	switch {
	case x.L == nil && x.R == nil:
		if x.Root() {
			t.root = nil
			return
		}

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

loop:
	if x.Root() {
		return
	}

	if x.Red() || x.P.Red() {
		x.P.color = black
		return
	}

	s := x.S()
	if x.Balck() && x.P.Balck() && s != nil {
		if s.Balck() {
			if (s.L != nil && s.L.Red()) || (s.R != nil && s.R.Red()) {
			sw:
				switch {
				case s.Left() && (s.L != nil && s.L.Red()): // LL
					y := s.R
					if y != nil {
						y.P = x.P
					}
					x.P.L = y

					s.P = x.P.P
					if s.P == nil {
						t.root = s
					}

					s.R = x.P
					x.P.P = s

					s.L.color = black

				case s.Left() && (s.R != nil && s.R.Red()): // LR
					s.color = red
					s.R.color = black

					s.R.P = x.P
					s.R.L = s

					s.P = s.R
					s.R = nil

					goto sw

				case s.Right() && (s.R != nil && s.R.Red()): // RR
					y := s.L
					if y != nil {
						y.P = x.P
					}
					x.P.R = y

					s.P = x.P.P
					if s.P == nil {
						t.root = s
					}

					s.L = x.P
					x.P.P = s

					s.R.color = black

				case s.Right() && (s.L != nil && s.L.Red()): // RL
					s.color = red
					s.L.color = black

					s.L.P = x.P
					s.L.R = s

					s.P = s.L
					s.L = nil

					goto sw
				}

			} else if (s.L == nil || s.L.Balck()) && (s.R == nil || s.R.Balck()) {
				s.color = red
				if x.P.Red() {
					x.P.color = black
					return
				}

				x = x.P
				goto loop
			}

		} else {

		}
	}
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
