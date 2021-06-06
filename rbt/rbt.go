package rbt

import (
	"container/list"
	"fmt"
	"strconv"
)

type color uint8

const (
	red   color = 1
	black color = 2
)

type node struct {
	color   color
	val     int
	L, R, P *node
}

func (n *node) String() string {
	if n.Red() {
		return strconv.Itoa(n.val) + "r"

	} else {
		return strconv.Itoa(n.val) + "b"
	}
}

func (n *node) Red() bool {
	return n.color == red
}

func (n *node) Balck() bool {
	return n.color == black
}

func (n *node) Root() bool {
	return n.P == nil
}

func (n *node) Left() bool {
	return n.P.L == n
}

func (n *node) Right() bool {
	return n.P.R == n
}

func (n *node) G() *node {
	if n.P == nil {
		return nil
	}
	return n.P.P
}

func (n *node) U() *node {
	g := n.G()
	if g == nil {
		return nil
	}

	if g.L == n.P {
		return g.R
	}
	return g.L
}

func (n *node) S() *node {
	p := n.P
	if p == nil {
		return nil
	}

	if p.L == n {
		return p.R
	}
	return p.L
}

func NewRbTree() *rbTree {
	return new(rbTree)
}

type rbTree struct {
	root *node
}

func (t *rbTree) Root() int {
	return t.root.val
}

func (t *rbTree) Asc() {
	var values []int
	push := func(value int) {
		values = append(values, value)
	}

	asc(t.root, push)

	// var values []int
	// iterator := t.Iterator()
	// for iterator.Next() {
	// 	values = append(values, iterator.Value())
	// }
	// fmt.Println(values)

	// values := t.InOrderNoRecursion()
	for i := 1; i < len(values); i++ {
		// if values[i]-values[i-1] != 1 {
		// 	fmt.Println(values[:i+1])
		// 	panic("xxx")
		// }

		if values[i] < values[i-1] {
			fmt.Println(values[:i+1])
			panic("xxx")
		}
	}
}

func asc(root *node, push func(value int)) {
	if root != nil {
		asc(root.L, push)
		push(root.val)
		asc(root.R, push)
	}
}

func (t *rbTree) InOrderNoRecursion() []int {
	root := t.root
	stack := list.New()
	res := make([]int, 0)
	for root != nil || stack.Len() != 0 {
		for root != nil {
			stack.PushBack(root)
			root = root.L
		}
		if stack.Len() != 0 {
			v := stack.Back()
			root = v.Value.(*node)
			res = append(res, root.val) //visit
			root = root.R
			stack.Remove(v)
		}
	}
	return res
}

func reverse(a []int) {
	for i, n := 0, len(a); i < n/2; i++ {
		a[i], a[n-1-i] = a[n-1-i], a[i]
	}
}

func postorderTraversal(root *node) (res []int) {
	addPath := func(node *node) {
		resSize := len(res)
		for ; node != nil; node = node.R {
			res = append(res, node.val)
		}
		reverse(res[resSize:])
	}

	p1 := root
	for p1 != nil {
		if p2 := p1.L; p2 != nil {
			for p2.R != nil && p2.R != p1 {
				p2 = p2.R
			}
			if p2.R == nil {
				p2.R = p1
				p1 = p1.L
				continue
			}
			p2.R = nil
			addPath(p1.L)
		}
		p1 = p1.R
	}
	addPath(root)
	return
}

func (t *rbTree) String() string {
	str := "RedBlackTree\n"
	if t.root != nil {
		output(t.root, "", true, &str)
	}
	return str
}

func output(root *node, prefix string, isTail bool, str *string) {
	if root.R != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "│   "
		} else {
			newPrefix += "    "
		}
		output(root.R, newPrefix, false, str)
	}
	*str += prefix
	if isTail {
		*str += "└── "
	} else {
		*str += "┌── "
	}
	*str += root.String() + "\n"
	if root.L != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		output(root.L, newPrefix, true, str)
	}
}
