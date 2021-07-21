package bpt

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

	cur := t.root
	for {
		index, bingo := search(cur, val)
		if !bingo {
			// TODO optimize, check the min & max value firstly
			if len(cur.children) == 0 { // leaf
				// not exists
				return
			}

			cur = cur.children[index]
			continue
		}

		if !cur.leaf() { // node
			if cur.overHalf() {
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

					cur = x
					continue
				}
			}
		}

		{ // leaf
			if cur.overHalf() {
				cur.values = append(cur.values[:index], cur.values[index+1:]...)
				return
			}
		}
	}
}
