package bpt

// import (
// 	"fmt"
// )

func (t *bpTree) Delete(val Value) {
	if val == nil {
		return
	}

	t.Lock()
	defer t.Unlock()

	t.delete(val)
}

func (t *bpTree) delete(val Value) {
	if val == nil {
		return
	}

	var (
		cur          = t.root
		parent       *node
		parentCindex int
	)

	var sL, sR *node
	rotate2Left := func() {
		cur.values = append(cur.values, parent.values[parentCindex])
		parent.values[parentCindex] = sR.values[0]
		sR.values = sR.values[1:]

		if !sR.leaf() {
			cur.children = append(cur.children, sR.children[0])
			sR.children = sR.children[1:]
		}
	}

	merge2Left := func() {
		cur.values = append(append(cur.values, parent.values[parentCindex]), sR.values...)
		cur.children = append(cur.children, sR.children...)

		if len(parent.values) == 1 {
			t.root = cur
		} else {
			parent.values = append(parent.values[:parentCindex], parent.values[parentCindex+1:]...)
			parent.children = append(parent.children[:parentCindex+1], parent.children[parentCindex+2:]...)
		}
	}

	rotate2Right := func() {
		cur.values = append(cur.values, cur.values[0])
		copy(cur.values[1:], cur.values)
		cur.values[0] = parent.values[parentCindex-1]

		parent.values[parentCindex-1] = sL.values[len(sL.values)-1]
		sL.values = sL.values[:len(sL.values)-1]

		if !sL.leaf() {
			cur.children = append(cur.children, cur.children[0])
			copy(cur.children[1:], cur.children)
			cur.children[0] = sL.children[len(sL.children)-1]

			sL.children = sL.children[:len(sL.children)-1]
		}
	}

	merge2Right := func() {
		cur.values = append(append(sL.values, parent.values[parentCindex-1]), cur.values...)
		cur.children = append(sL.children, cur.children...)

		if len(parent.values) == 1 {
			t.root = cur
		} else {
			parent.values = append(parent.values[:parentCindex-1], parent.values[parentCindex:]...)
			parent.children = append(parent.children[:parentCindex-1], parent.children[parentCindex:]...)
		}
	}

	for {
		// internal node and half
		if cur != t.root && !cur.overHalf() {
			switch {
			case parentCindex == 0:
				if sR = parent.children[parentCindex+1]; sR.overHalf() {
					rotate2Left()

				} else {
					merge2Left()
				}

			case parentCindex == len(parent.values):
				if sL = parent.children[parentCindex-1]; sL.overHalf() {
					rotate2Right()

				} else {
					merge2Right()
				}

			default:
				sL = parent.children[parentCindex-1]
				sR = parent.children[parentCindex+1]

				if sL.overHalf() {
					rotate2Right()

				} else if sR.overHalf() {
					rotate2Left()

				} else {
					merge2Left()
				}
			}
			continue
		}

		index, bingo := search(cur, val)
		if !bingo {
			// TODO optimize, check the min & max value firstly
			if cur.leaf() {
				// not exists
				return
			}

			parent, parentCindex = cur, index
			cur = cur.children[index]
			continue
		}

		if !cur.leaf() { // node
			{ // left child over half
				x := cur.children[index]
				if x.overHalf() {
					val = x.values[len(x.values)-1]
					cur.values[index] = val
					cur = x
					continue
				}
			}

			{ // right child over half
				y := cur.children[index+1]
				if y.overHalf() {
					val = y.values[0]
					cur.values[index] = val
					cur = y
					continue
				}
			}

			{ // both left and right are half
				x := cur.children[index]
				y := cur.children[index+1]

				x.values = append(x.values, cur.values[index])
				x.values = append(x.values, y.values...)
				x.children = append(x.children, y.children...)

				cur.values = append(cur.values[:index], cur.values[index+1:]...)
				cur.children = append(cur.children[:index+1], cur.children[index+2:]...)

				if cur == t.root && len(cur.values) == 0 {
					t.root = x
				}
				cur = x
				continue
			}
		}

		{ // leaf
			if cur.overHalf() || cur == t.root {
				cur.values = append(cur.values[:index], cur.values[index+1:]...)
				return
			}

			switch {
			case parentCindex == 0:
				if sR = parent.children[parentCindex+1]; sR.overHalf() {
					rotate2Left()

				} else {
					merge2Left()
				}

			case parentCindex == len(parent.values):
				if sL = parent.children[parentCindex-1]; sL.overHalf() {
					rotate2Right()

				} else {
					merge2Right()
				}

			default:
				sL = parent.children[parentCindex-1]
				sR = parent.children[parentCindex+1]

				if sL.overHalf() {
					rotate2Right()

				} else if sR.overHalf() {
					rotate2Left()

				} else {
					merge2Left()
				}
			}
		}
	}
}
