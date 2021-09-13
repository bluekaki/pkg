package bpt

import (
	"github.com/bluekaki/pkg/stringutil"
)

func (t *bpTree) Delete(val Value) (ok bool) {
	if val == nil {
		return
	}

	t.Lock()
	defer t.Unlock()

	return t.delete(val)
}

func (t *bpTree) delete(val Value) (ok bool) {
	if val == nil || t.size == 0 {
		return
	}

	var (
		cur = t.root

		parent = &struct {
			*node
			cIndex int
		}{}

		toReplaced = &struct {
			target *node
			val    Value
		}{}
	)

	var sL, sR *node
	rotate2Left := func() {
		if toReplaced.target == parent.node && toReplaced.val.Compare(parent.values[parent.cIndex]) == stringutil.Equal {
			toReplaced.target = cur
		}

		cur.values = append(cur.values, parent.values[parent.cIndex])
		parent.values[parent.cIndex] = sR.values[0]
		sR.values = sR.values[1:]

		if !sR.leaf() {
			cur.children = append(cur.children, sR.children[0])
			sR.children = sR.children[1:]
		}
	}

	merge2Left := func() {
		if toReplaced.target == parent.node && toReplaced.val.Compare(parent.values[parent.cIndex]) == stringutil.Equal {
			toReplaced.target = cur
		}

		cur.values = append(append(cur.values, parent.values[parent.cIndex]), sR.values...)
		cur.children = append(cur.children, sR.children...)

		if len(parent.values) == 1 {
			t.root = cur
		} else {
			parent.values = append(parent.values[:parent.cIndex], parent.values[parent.cIndex+1:]...)
			parent.children = append(parent.children[:parent.cIndex+1], parent.children[parent.cIndex+2:]...)
		}
	}

	rotate2Right := func() {
		if toReplaced.target == parent.node && toReplaced.val.Compare(parent.values[parent.cIndex-1]) == stringutil.Equal {
			toReplaced.target = cur
		}

		cur.values = append(cur.values, cur.values[0])
		copy(cur.values[1:], cur.values)
		cur.values[0] = parent.values[parent.cIndex-1]

		parent.values[parent.cIndex-1] = sL.values[len(sL.values)-1]
		sL.values = sL.values[:len(sL.values)-1]

		if !sL.leaf() {
			cur.children = append(cur.children, cur.children[0])
			copy(cur.children[1:], cur.children)
			cur.children[0] = sL.children[len(sL.children)-1]

			sL.children = sL.children[:len(sL.children)-1]
		}
	}

	merge2Right := func() {
		if toReplaced.target == parent.node && toReplaced.val.Compare(parent.values[parent.cIndex-1]) == stringutil.Equal {
			toReplaced.target = cur
		}

		cur.values = append(append(sL.values, parent.values[parent.cIndex-1]), cur.values...)
		cur.children = append(sL.children, cur.children...)

		if len(parent.values) == 1 {
			t.root = cur
		} else {
			parent.values = append(parent.values[:parent.cIndex-1], parent.values[parent.cIndex:]...)
			parent.children = append(parent.children[:parent.cIndex-1], parent.children[parent.cIndex:]...)
		}
	}

	for {
		// internal node and half
		if cur != t.root && !cur.overHalf(t.meta.HT) {
			switch {
			case parent.cIndex == 0:
				if sR = parent.children[parent.cIndex+1]; sR.overHalf(t.meta.HT) {
					rotate2Left()

				} else {
					merge2Left()
				}

			case parent.cIndex == len(parent.values):
				if sL = parent.children[parent.cIndex-1]; sL.overHalf(t.meta.HT) {
					rotate2Right()

				} else {
					merge2Right()
				}

			default:
				sL = parent.children[parent.cIndex-1]
				sR = parent.children[parent.cIndex+1]

				if sL.overHalf(t.meta.HT) {
					rotate2Right()

				} else if sR.overHalf(t.meta.HT) {
					rotate2Left()

				} else {
					merge2Left()
				}
			}
			continue
		}

		index, bingo := t.search(cur, val)
		if !bingo {
			if cur.leaf() {
				// not exists
				return
			}
			parent.node, parent.cIndex = cur, index

			cur = cur.children[index]
			continue
		}

		if !cur.leaf() { // node
			if toReplaced.target == nil {
				x := cur.children[index]

				toReplaced.target = cur
				toReplaced.val = val

				parent.node, parent.cIndex = cur, index
				val = x.values[len(x.values)-1]
				cur = x

			} else {
				y := cur.children[index+1]

				parent.node, parent.cIndex = cur, index+1
				val = y.values[len(y.values)-1]
				cur = y
			}
			continue
		}

		{ // leaf
			cur.values = append(cur.values[:index], cur.values[index+1:]...)
			if toReplaced.target != nil {
				index, _ = t.search(toReplaced.target, toReplaced.val)
				toReplaced.target.values[index] = val
			}

			if t.size--; t.size == 0 {
				t.root = nil
			}
			ok = true
			return
		}
	}
}
