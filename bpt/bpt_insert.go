package bpt

import (
	"sort"

	"github.com/bluekaki/pkg/stringutil"
)

func (t *bpTree) Add(val Value) (ok bool) {
	if val == nil {
		return
	}

	t.Lock()
	defer t.Unlock()

	if t.root == nil {
		t.root = &node{
			id:     ID(),
			values: []Value{val},
		}

		t.size++
		ok = true
		return
	}

	if t.root.full() {
		x, y, mid := split(t.root)

		t.root.values = []Value{mid}
		t.root.children = []*node{x, y}
	}

	cur := t.root
	for {
		if !cur.leaf() { // node
			index, duplicated := search(cur, val)
			if duplicated {
				cur.values[index] = val
				return
			}

			if cur.children[index].full() {
				x, y, mid := split(cur.children[index])

				cur.values = append(cur.values, cur.values[0])
				copy(cur.values[index+1:], cur.values[index:])
				cur.values[index] = mid

				cur.children = append(cur.children, cur.children[0])
				copy(cur.children[index+2:], cur.children[index+1:])
				cur.children[index] = x
				cur.children[index+1] = y

			} else {
				cur = cur.children[index]
			}
			continue
		}

		{ // leaf
			index, duplicated := search(cur, val)
			if duplicated {
				cur.values[index] = val
				return
			}

			cur.values = append(cur.values, cur.values[0])
			copy(cur.values[index+1:], cur.values[index:])
			cur.values[index] = val

			t.size++
			ok = true
			return
		}
	}
}

func search(cur *node, val Value) (index int, duplicated bool) {
	switch cur.values[0].Compare(val) {
	case stringutil.Equal:
		return 0, true

	case stringutil.Greater:
		return 0, false
	}

	switch cur.values[len(cur.values)-1].Compare(val) {
	case stringutil.Equal:
		return len(cur.values) - 1, true

	case stringutil.Less:
		return len(cur.values), false
	}

	index = sort.Search(len(cur.values), func(i int) bool {
		diff := cur.values[i].Compare(val)
		return diff == stringutil.Greater || diff == stringutil.Equal
	})

	if index < len(cur.values) {
		duplicated = cur.values[index].Compare(val) == stringutil.Equal
	}
	return
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
		id:       ID(),
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
		id:       ID(),
		values:   yV,
		children: yC,
	}

	mid = cur.values[_Mid]
	return
}
