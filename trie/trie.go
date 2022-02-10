package trie

import (
	"bytes"
	"container/list"
)

type node struct {
	char rune
	next []*node
}

func (n *node) leaf() bool {
	return len(n.next) == 0
}

func (n *node) search(char rune) (index int, ok bool) {
	for i, n := range n.next {
		if n.char == char {
			return i, true
		}
	}

	return -1, false
}

func (n *node) insert(char rune) (*node, bool) {
	cur := n
	for {
		if cur.char == char {
			return cur, false
		}

		index, ok := cur.search(char)
		if ok {
			cur = cur.next[index]
			continue
		}

		cur.next = append(cur.next, &node{char: char})
		return cur.next[len(cur.next)-1], true
	}
}

func (n *node) delete(index int) {
	if len(n.next) != 0 {
		n.next = append(n.next[:index], n.next[index+1:]...)
	}
}

type Trie struct {
	root *node
	size uint32
}

func (t *Trie) String() string {
	buf := bytes.NewBuffer(nil)
	for _, cur := range t.root.next {
		for _, phrase := range walkNode(cur) {
			buf.WriteString(phrase)
			buf.WriteString("\n")
		}
	}

	msg := buf.String()
	return msg[:len(msg)-1] // trim \n
}

func walkNode(cur *node) (phrases []string) {
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
			var chars []rune
			for itor := stack.Front(); itor != nil; itor = itor.Next() {
				chars = append(chars, itor.Value.(*Entry).node.char)
			}
			chars = append(chars, entry.node.char)

			phrases = append(phrases, string((chars)))
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

func (t *Trie) Insert(val string) {
	if t.root == nil {
		t.root = new(node)
	}

	var cur *node
	chars := []rune(val)

	index, ok := t.root.search(chars[0])
	if ok {
		cur = t.root.next[index]

	} else {
		t.root.next = append(t.root.next, &node{char: chars[0]})
		t.size++

		cur = t.root.next[len(t.root.next)-1]
	}

	for _, char := range chars[1:] {
		cur, ok = cur.insert(char)
		if ok {
			t.size++
		}
	}
}

func (t *Trie) Exists(val string) bool {
	cur := t.root
	for _, char := range []rune(val) {
		index, ok := cur.search(char)
		if !ok {
			return false
		}

		cur = cur.next[index]
	}

	return true
}

func (t *Trie) Delete(val string) {
	type Entry struct {
		node  *node
		index int
	}
	var path []*Entry

	cur := t.root
	for _, char := range []rune(val) {
		index, ok := cur.search(char)
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
	t.size--

	for k := len(path) - 2; k >= 0; k-- { // the second last
		if cur := path[k]; last.node.leaf() {
			cur.node.delete(cur.index)
			t.size--

			last = cur
			continue
		}
		return
	}
}

func (t *Trie) Prompt(prefix string) []string {

	return nil
}
