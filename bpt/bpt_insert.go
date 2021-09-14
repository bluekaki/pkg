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

	if !fileExists(rootIndex, t.meta.baseDir) {
		(&node{
			index:  rootIndex,
			values: []Value{val},
		}).takeSnapshots(t.meta.baseDir, t.logger)

		t.size++
		return true
	}

	root := t.loadSnapshots(rootIndex)
	if root.full(t.meta.N) {
		x, y, mid := t.split(root)
		x.takeSnapshots(t.meta.baseDir, t.logger)
		y.takeSnapshots(t.meta.baseDir, t.logger)

		root.values = []Value{mid}
		root.children = []*node{x, y}
		root.takeSnapshots(t.meta.baseDir, t.logger)
	}

	cur := root
	for {
		if !cur.leaf() { // node
			index, duplicated := t.search(cur, val)
			if duplicated {
				cur.values[index] = val
				cur.takeSnapshots(t.meta.baseDir, t.logger)
				return
			}

			if child := t.loadSnapshots(cur.children[index].index); child.full(t.meta.N) {
				x, y, mid := t.split(child)
				child.delete(t.meta.baseDir, t.logger)
				x.takeSnapshots(t.meta.baseDir, t.logger)
				y.takeSnapshots(t.meta.baseDir, t.logger)

				cur.values = append(cur.values, cur.values[0])
				copy(cur.values[index+1:], cur.values[index:])
				cur.values[index] = mid

				cur.children = append(cur.children, cur.children[0])
				copy(cur.children[index+2:], cur.children[index+1:])
				cur.children[index] = x
				cur.children[index+1] = y

				cur.takeSnapshots(t.meta.baseDir, t.logger)

			} else {
				cur = t.loadSnapshots(cur.children[index].index)
			}
			continue
		}

		{ // leaf
			index, duplicated := t.search(cur, val)
			if duplicated {
				cur.values[index] = val
				cur.takeSnapshots(t.meta.baseDir, t.logger)
				return
			}

			cur.values = append(cur.values, cur.values[0])
			copy(cur.values[index+1:], cur.values[index:])
			cur.values[index] = val
			cur.takeSnapshots(t.meta.baseDir, t.logger)

			t.size++
			return true
		}
	}
}

func (t *bpTree) search(cur *node, val Value) (index int, duplicated bool) {
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

func (t *bpTree) split(cur *node) (x, y *node, mid Value) {
	xV := make([]Value, len(cur.values[:t.meta.Mid]))
	copy(xV, cur.values[:t.meta.Mid])

	var xC []*node
	if len(cur.children) > 0 {
		xC = make([]*node, len(cur.children[:t.meta.HT]))
		copy(xC, cur.children[:t.meta.HT])
	}

	x = &node{
		index:    t.nextIndex(),
		values:   xV,
		children: xC,
	}

	yV := make([]Value, len(cur.values[t.meta.Mid+1:]))
	copy(yV, cur.values[t.meta.Mid+1:])

	var yC []*node
	if len(cur.children) > 0 {
		yC = make([]*node, len(cur.children[t.meta.HT:]))
		copy(yC, cur.children[t.meta.HT:])
	}

	y = &node{
		index:    t.nextIndex(),
		values:   yV,
		children: yC,
	}

	mid = cur.values[t.meta.Mid]
	return
}
