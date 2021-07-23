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
		parent       *node
		parentCindex int
	)

	cur := t.root
	for {
		index, bingo := search(cur, val)
		if !bingo {
			// TODO optimize, check the min & max value firstly
			if len(cur.children) == 0 { // leaf
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

				if cur == t.root {
					t.root = x
				}
				cur = x
				continue
			}
		}

		{ // leaf
			// fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
			// fmt.Println(t.root)
			// fmt.Println(cur)
			// fmt.Println("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")

			if cur.overHalf() || cur == t.root {
				cur.values = append(cur.values[:index], cur.values[index+1:]...)
				return
			}

			var sL, sR *node
			rotate2Left := func() {
				cur.values = append(cur.values, parent.values[parentCindex])
				parent.values[parentCindex] = sR.values[0]
				sR.values = sR.values[1:]
			}

			merge2Left := func() {
				cur.values = append(append(cur.values, parent.values[parentCindex]), sR.values...)

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
			}

			merge2Right := func() {
				cur.values = append(append(sL.values, parent.values[parentCindex-1]), cur.values...)

				if len(parent.values) == 1 {
					t.root = cur
				} else {
					parent.values = append(parent.values[:parentCindex-1], parent.values[parentCindex:]...)
					parent.children = append(parent.children[:parentCindex-1], parent.children[parentCindex:]...)
				}
			}

			switch {
			case parentCindex == 0:
				if sR = parent.children[parentCindex+1]; sR.overHalf() {
					// fmt.Println("------case0 rotate2Left-------")
					rotate2Left()
				} else {
					// fmt.Println("------case0 merge2Left-------")
					merge2Left()
				}

			case parentCindex == len(parent.values):
				if sL = parent.children[parentCindex-1]; sL.overHalf() {
					// fmt.Println("------case1 rotate2Right-------")
					rotate2Right()
				} else {
					// fmt.Println("------case1 merge2Right-------")
					merge2Right()
				}

			default:
				// fmt.Println("+++++++++++++++++++", parentCindex)
				sL = parent.children[parentCindex-1]
				sR = parent.children[parentCindex+1]

				if sL.overHalf() {
					// fmt.Println("------case2 rotate2Right-------")
					rotate2Right()
				} else if sR.overHalf() {
					// fmt.Println("------case2 rotate2Left-------")
					rotate2Left()
				} else {
					// fmt.Println("------case2 merge2Left-------")
					merge2Left()
				}
			}
		}
	}
}
