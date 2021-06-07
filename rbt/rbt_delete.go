package rbt

func (t *rbTree) Delete(val Value) {
	if val == nil {
		return
	}

	t.Lock()
	defer t.Unlock()

	t.delete(val)
}

func (t *rbTree) delete(val Value) {
	if val == nil {
		return
	}

	x := t.lookup(val)
	if x == nil {
		return
	}

	var ok bool
	if x.values, ok = delVal(x.values, val); !ok {
		return
	}

	t.size--
	if len(x.values) != 0 {
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
			x.values = s.values
			x = s

		} else {
			s := t.maximum(x.L)
			x.values = s.values
			x = s
		}

		goto del

	default:
		s := t.minimum(x.R)
		x.values = s.values
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
				switch {
				case s.Left() && (s.L != nil && s.L.Red()): // LL
					t.rotateToTheLeft(x, s)
					s.L.color = black

				case s.Left() && (s.R != nil && s.R.Red()): // LR
					s.color = red
					s.R.color = black

					s.R.P = x.P
					x.P.L = s.R

					y := s.R.L
					if y != nil {
						y.P = s
					}

					s.R.L = s
					s.P = s.R
					s.R = y

					goto loop

				case s.Right() && (s.R != nil && s.R.Red()): // RR
					t.rotateToTheRight(x, s)
					s.R.color = black

				case s.Right() && (s.L != nil && s.L.Red()): // RL
					s.color = red
					s.L.color = black

					s.L.P = x.P
					x.P.R = s.L

					y := s.L.R
					if y != nil {
						y.P = s
					}

					s.L.R = s
					s.P = s.L
					s.L = y

					goto loop
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
			s.color = black

			if s.Left() {
				t.rotateToTheLeft(x, s)

				if s.L != nil {
					s.L.color = red
				}

			} else {
				t.rotateToTheRight(x, s)

				if s.R != nil {
					s.R.color = red
				}
			}
		}
	}
}

func (t *rbTree) lookup(val Value) *node {
	root := t.root
	for root != nil {
		switch val.Compare(root.values[0]) {
		case Less:
			root = root.L

		case Greater:
			root = root.R

		case Equal:
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

func (t *rbTree) rotateToTheLeft(x, s *node) {
	y := s.R
	if y != nil {
		y.P = x.P
	}
	x.P.L = y

	s.P = x.P.P
	if s.P == nil {
		t.root = s
	} else {
		if x.P.Left() {
			x.P.P.L = s
		} else {
			x.P.P.R = s
		}
	}

	s.R = x.P
	x.P.P = s
}

func (t *rbTree) rotateToTheRight(x, s *node) {
	y := s.L
	if y != nil {
		y.P = x.P
	}
	x.P.R = y

	s.P = x.P.P
	if s.P == nil {
		t.root = s
	} else {
		if x.P.Left() {
			x.P.P.L = s
		} else {
			x.P.P.R = s
		}
	}

	s.L = x.P
	x.P.P = s
}

func delVal(values []Value, val Value) ([]Value, bool) {
	for i, value := range values {
		if value.ID() == val.ID() {
			return append(values[:i], values[i+1:]...), true
		}
	}

	return values, false
}
