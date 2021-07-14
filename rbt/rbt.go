package rbt

import (
	"bytes"
	"container/list"
	"sync"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/rbt/internal/pb/gen"

	"github.com/golang/protobuf/proto"
)

func NewRbTree() *rbTree {
	return new(rbTree)
}

type rbTree struct {
	sync.RWMutex
	size uint32
	root *node
}

func (t *rbTree) String() string {
	t.RLock()
	defer t.RUnlock()

	buf := bytes.NewBufferString("RedBlackTree\n")
	if t.root != nil {
		output(t.root, "", true, buf)
	}
	return buf.String()
}

func output(root *node, prefix string, isTail bool, buf *bytes.Buffer) {
	if root.R != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "│   "
		} else {
			newPrefix += "    "
		}

		output(root.R, newPrefix, false, buf)
	}

	buf.WriteString(prefix)
	if isTail {
		buf.WriteString("└── ")
	} else {
		buf.WriteString("┌── ")
	}

	buf.WriteString(root.String())
	buf.WriteString("\n")

	if root.L != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}

		output(root.L, newPrefix, true, buf)
	}
}

func (t *rbTree) Empty() bool {
	t.RLock()
	defer t.RUnlock()

	return t.size == 0
}

func (t *rbTree) Size() uint32 {
	t.RLock()
	defer t.RUnlock()

	return t.size
}

func (t *rbTree) Asc() []Value {
	t.RLock()
	defer t.RUnlock()

	if t.size == 0 {
		return nil
	}

	values := make([]Value, 0, t.size)

	stack := list.New()
	root := t.root
	for root != nil || stack.Len() != 0 {
		for root != nil {
			stack.PushBack(root)
			root = root.L
		}

		if stack.Len() != 0 {
			v := stack.Back()
			root = v.Value.(*node)
			values = append(values, root.values...) // visit

			root = root.R
			stack.Remove(v)
		}
	}

	return values
}

func (t *rbTree) Desc() []Value {
	t.RLock()
	defer t.RUnlock()

	if t.size == 0 {
		return nil
	}

	values := make([]Value, 0, t.size)

	stack := list.New()
	root := t.root
	for root != nil || stack.Len() != 0 {
		for root != nil {
			stack.PushBack(root)
			root = root.R
		}

		if stack.Len() != 0 {
			v := stack.Back()
			root = v.Value.(*node)
			values = append(values, root.values...) // visit

			root = root.L
			stack.Remove(v)
		}
	}

	return values
}

func (t *rbTree) Maximum() []Value {
	t.RLock()
	defer t.RUnlock()

	if t.root == nil {
		return nil
	}

	root := t.maximum(t.root)
	return root.values
}

func (t *rbTree) PopMaximum() []Value {
	t.Lock()
	defer t.Unlock()

	if t.root == nil {
		return nil
	}

	root := t.maximum(t.root)
	values := make([]Value, len(root.values))
	copy(values, root.values)

	for _, val := range values {
		t.delete(val)
	}
	return values
}

func (t *rbTree) Minimum() []Value {
	t.RLock()
	defer t.RUnlock()

	if t.root == nil {
		return nil
	}

	root := t.minimum(t.root)
	return root.values
}

func (t *rbTree) PopMinimum() []Value {
	t.Lock()
	defer t.Unlock()

	if t.root == nil {
		return nil
	}

	root := t.minimum(t.root)
	values := make([]Value, len(root.values))
	copy(values, root.values)

	for _, val := range values {
		t.delete(val)
	}
	return values
}

func (t *rbTree) Marshal() []byte {
	t.Lock()
	defer t.Unlock()

	if t.root == nil {
		return nil
	}

	nodes := make([]*node, 0, t.size)
	curLayer := []*node{t.root}

	for len(curLayer) > 0 {
		var nexLayer []*node

		for _, node := range curLayer {
			nodes = append(nodes, node)
			if node.L != nil {
				nexLayer = append(nexLayer, node.L)
			}
			if node.R != nil {
				nexLayer = append(nexLayer, node.R)
			}
		}

		curLayer = nexLayer
	}

	tree := &pb.Tree{
		Nodes: make([]*pb.Tree_Node, 0, t.size),
	}
	for _, node := range nodes {
		for _, value := range node.values {
			tree.Nodes = append(tree.Nodes, &pb.Tree_Node{
				Val: value.Marshal(),
			})
		}
	}

	raw, _ := proto.Marshal(tree)
	return raw
}

func Unmarshal(payload []byte, unmarshaler func(raw []byte) Value) (*rbTree, error) {
	if unmarshaler == nil {
		return nil, errors.New("unmarshaler required")
	}

	if len(payload) == 0 {
		return new(rbTree), nil
	}

	data := new(pb.Tree)
	err := proto.Unmarshal(payload, data)
	if err != nil {
		return nil, errors.Wrap(err, "payload illegal")
	}

	rbt := new(rbTree)
	for _, node := range data.Nodes {
		rbt.Add(unmarshaler(node.Val))
	}

	return rbt, nil
}
