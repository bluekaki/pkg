package rbt

import (
	"fmt"
	"os"
	// "sync/atomic"
)

// var counter uint64

func (t *rbTree) Delete(value int, file *os.File) {
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

		} else if x.Right() {
			x.P.R = nil
		}

		if true {
			file.WriteString(fmt.Sprintf("-- %v %v %v %v --\n", x.val, x.P.val, x.Left(), x.Right()))
			file.WriteString(fmt.Sprintf("left: %+v\n", x.P.L))
			file.WriteString(fmt.Sprintf("left: %+v\n", x.P.R))
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

		// fmt.Println("goto del 1", atomic.AddUint64(&counter, 1))
		goto del

	default:
		s := t.minimum(x.R)
		x.val = s.val
		x = s

		// fmt.Println("goto del 2", atomic.AddUint64(&counter, 1))
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
					} else {
						if x.P.Left() {
							x.P.P.L = s
						} else {
							x.P.P.R = s
						}
					}

					s.R = x.P
					x.P.P = s

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

					// fmt.Println("goto sw 1", atomic.AddUint64(&counter, 1))
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
					} else {
						if x.P.Left() {
							x.P.P.L = s
						} else {
							x.P.P.R = s
						}
					}

					s.L = x.P
					x.P.P = s

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

					// fmt.Println("goto sw 2", atomic.AddUint64(&counter, 1))
					goto sw
				}

			} else if (s.L == nil || s.L.Balck()) && (s.R == nil || s.R.Balck()) {
				s.color = red
				if x.P.Red() {
					x.P.color = black
					return
				}

				x = x.P

				// fmt.Println("goto loop 1", atomic.AddUint64(&counter, 1))
				goto loop
			}

		} else {
			s.color = black

			if s.Left() {
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

				if s.L != nil {
					s.L.color = red
				}

			} else {
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

				if s.R != nil {
					s.R.color = red
				}
			}
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
