package trie

import (
	"bytes"
	"container/list"
	"strings"
	"sync"
)

const (
	EmptyDelimiter     = ""
	SpaceDelimiter     = " "
	SlashDelimiter     = "/"
	BackSlashDelimiter = "\\"
)

type Trie interface {
	t()
	Capacity() uint32
	String(delimiter string) string
	Insert(values []string)
	Exists(values []string) bool
	Delete(values []string)
	Prompt(prefix []string, delimiter string) (phrases []string)
}

type node struct {
	val  string
	next []*node
}

func (n *node) leaf() bool {
	return len(n.next) == 0
}

func (n *node) search(val string) (index int, ok bool) {
	for i, n := range n.next {
		if n.val == val {
			return i, true
		}
	}

	return -1, false
}

func (n *node) insert(val string) (*node, bool) {
	cur := n
	for {
		if cur.val == val {
			return cur, false
		}

		index, ok := cur.search(val)
		if ok {
			cur = cur.next[index]
			continue
		}

		cur.next = append(cur.next, &node{val: val})
		return cur.next[len(cur.next)-1], true
	}
}

func (n *node) delete(index int) {
	if len(n.next) != 0 {
		n.next = append(n.next[:index], n.next[index+1:]...)
	}
}

func New() Trie {
	return &trie{root: new(node)}
}

type trie struct {
	sync.RWMutex
	root     *node
	capacity uint32
}

func (t *trie) t() {}

func (t *trie) Capacity() uint32 {
	t.RLock()
	defer t.RUnlock()

	return t.capacity
}

func (t *trie) String(delimiter string) string {
	t.RLock()
	defer t.RUnlock()

	buf := bytes.NewBuffer(nil)
	for _, cur := range t.root.next {
		for _, phrase := range walkNode(cur, delimiter) {
			buf.WriteString(phrase)
			buf.WriteString("\n")
		}
	}

	msg := buf.String()
	if len(msg) > 0 {
		msg = msg[:len(msg)-1] // trim \n
	}

	return msg
}

func walkNode(cur *node, delimiter string) (phrases []string) {
	type Entry struct {
		node  *node
		index int
	}

	stack := list.New()
	stack.PushBack(&Entry{node: cur})

	for stack.Len() > 0 {
		element := stack.Back()
		stack.Remove(element)

		entry := element.Value.(*Entry)
		if entry.node.leaf() {
			var values []string
			for itor := stack.Front(); itor != nil; itor = itor.Next() {
				values = append(values, itor.Value.(*Entry).node.val)
			}
			values = append(values, entry.node.val)

			phrases = append(phrases, strings.Join(values, delimiter))
			continue
		}

		if entry.index < len(entry.node.next) {
			next := entry.node.next[entry.index]

			entry.index++
			stack.PushBack(entry)
			stack.PushBack(&Entry{node: next})
		}
	}

	return
}

func (t *trie) Insert(values []string) {
	t.Lock()
	defer t.Unlock()

	cur := t.root
	var ok bool

	for _, val := range values {
		cur, ok = cur.insert(val)
		if ok {
			t.capacity++
		}
	}
}

func (t *trie) Exists(values []string) bool {
	t.RLock()
	defer t.RUnlock()

	cur := t.root
	for _, val := range values {
		index, ok := cur.search(val)
		if !ok {
			return false
		}

		cur = cur.next[index]
	}

	return true
}

func (t *trie) Delete(values []string) {
	t.Lock()
	defer t.Unlock()

	type Entry struct {
		node  *node
		index int
	}
	var path []*Entry

	cur := t.root
	for _, val := range values {
		index, ok := cur.search(val)
		if !ok {
			return
		}

		path = append(path, &Entry{node: cur, index: index})
		cur = cur.next[index]
	}

	if !cur.leaf() { // not leaf
		return
	}

	last := path[len(path)-1]
	last.node.delete(last.index)
	t.capacity--

	for k := len(path) - 2; k >= 0; k-- { // the second last
		if cur := path[k]; last.node.leaf() {
			cur.node.delete(cur.index)
			t.capacity--

			last = cur
			continue
		}

		return
	}
}

func (t *trie) Prompt(prefix []string, delimiter string) (phrases []string) {
	t.RLock()
	defer t.RUnlock()

	cur := t.root
	for _, val := range prefix {
		index, ok := cur.search(val)
		if !ok {
			return nil
		}

		cur = cur.next[index]
	}

	for _, next := range cur.next {
		phrases = append(phrases, walkNode(next, delimiter)...)
	}
	return
}
