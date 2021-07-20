package bpt

import (
	"sort"

	"github.com/bluekaki/pkg/stringutil"
)

func (t *bpTree) Add(val Value) {
	if val == nil {
		return
	}

	t.Lock()
	defer t.Unlock()

	if t.root == nil {
		t.root = &node{
			values: []Value{val},
		}
		t.size++
		return
	}

	cur := t.root
	for {
		if cur.full() {
			x, y, mid := split(cur)

			cur.values = []Value{mid}
			cur.children = []*node{x, y}
		}

		if len(cur.children) == 0 { // leaf
			{
				switch cur.values[0].Compare(val) {
				case stringutil.Equal:
					cur.values[0] = val
					return

				case stringutil.Greater:
					cur.values = append([]Value{val}, cur.values...)
					t.size++
					return
				}
			}

			{
				index := len(cur.values) - 1
				switch cur.values[index].Compare(val) {
				case stringutil.Equal:
					cur.values[index] = val
					return

				case stringutil.Less:
					cur.values = append(cur.values, val)
					t.size++
					return
				}
			}

			index := sort.Search(len(cur.values), func(i int) bool {
				diff := cur.values[i].Compare(val)
				return diff == stringutil.Greater || diff == stringutil.Equal
			})

			if cur.values[index].Compare(val) == stringutil.Equal { // duplicated
				cur.values[index] = val
				return
			}

			values := make([]Value, len(cur.values)+1)
			copy(values[:index], cur.values[:index])
			values[index] = val
			copy(values[index+1:], cur.values[index:])

			cur.values = values
			t.size++
			return
		}

		if len(cur.children) > 0 { // node
			{
				switch cur.values[0].Compare(val) {
				case stringutil.Equal:
					cur.values[0] = val
					return

				case stringutil.Greater:
					if cur.children[0].full() {
						x, y, mid := split(cur.children[0])
						cur.values = append([]Value{mid}, cur.values...)

						children := make([]*node, len(cur.children)+1)
						children[0] = x
						children[1] = y
						copy(children[2:], cur.children[1:])
						cur.children = children

					} else {
						cur = cur.children[0]
					}
					continue
				}
			}

			{
				index := len(cur.values) - 1
				switch cur.values[index].Compare(val) {
				case stringutil.Equal:
					cur.values[index] = val
					return

				case stringutil.Less:
					if cur.children[len(cur.children)-1].full() {
						x, y, mid := split(cur.children[len(cur.children)-1])
						cur.values = append(cur.values, mid)

						cur.children[len(cur.children)-1] = x
						cur.children = append(cur.children, y)

					} else {
						cur = cur.children[len(cur.children)-1]
					}
					continue
				}
			}

			index := sort.Search(len(cur.values), func(i int) bool {
				diff := cur.values[i].Compare(val)
				return diff == stringutil.Greater || diff == stringutil.Equal
			})

			if cur.values[index].Compare(val) == stringutil.Equal { // duplicated
				cur.values[index] = val
				return
			}

			if cur.children[index].full() {
				x, y, mid := split(cur.children[index])

				values := make([]Value, len(cur.values)+1)
				copy(values[:index], cur.values[:index])
				values[index] = mid
				copy(values[index+1:], cur.values[index:])
				cur.values = values

				children := make([]*node, len(cur.children)+1)
				copy(children[:index], cur.children[:index])
				children[index] = x
				children[index+1] = y
				copy(children[index+2:], cur.children[index+1:])
				cur.children = children

			} else {
				cur = cur.children[index]
			}
			continue
		}
	}
}

func split(cur *node) (x, y *node, mid Value) {
	xV := make([]Value, len(cur.values[:_Mid]))
	copy(xV, cur.values[:_Mid])

	var xC []*node
	if len(cur.children) > 0 {
		xC = make([]*node, len(cur.children[:_T]))
		copy(xC, cur.children[:_T])
	}

	x = &node{
		values:   xV,
		children: xC,
	}

	yV := make([]Value, len(cur.values[_Mid+1:]))
	copy(yV, cur.values[_Mid+1:])

	var yC []*node
	if len(cur.children) > 0 {
		yC = make([]*node, len(cur.children[_T:]))
		copy(yC, cur.children[_T:])
	}

	y = &node{
		values:   yV,
		children: yC,
	}

	mid = cur.values[_Mid]
	return
}
