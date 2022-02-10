package trie

import (
	"bytes"
	"container/list"
	"sort"
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
	HasPrefix(values []string) bool
	Match(values []string) bool
	Delete(values []string)
	Prompt(prefix []string, delimiter string) (phrases []string)
}

func SplitByEmpty(val string) []string {
	raw := []rune(val)
	values := make([]string, len(raw))
	for i, char := range raw {
		values[i] = string(char)
	}

	return values
}

func SplitByDelimiter(val string, delimiter string) []string {
	raw := strings.Split(strings.TrimLeft(strings.TrimSpace(val), delimiter), delimiter)
	values := make([]string, len(raw))
	for i, char := range raw {
		values[i] = char
	}

	return values
}

type node struct {
	val     string
	section bool
	next    []*node
}

func (n *node) leaf() bool {
	return len(n.next) == 0
}

func (n *node) search(val string) (int, bool) {
	if len(n.next) == 0 {
		return -1, false
	}

	index := sort.Search(len(n.next), func(i int) bool {
		return n.next[i].val >= val
	})

	if index >= len(n.next) || n.next[index].val != val {
		return -1, false
	}

	return index, true
}

func (n *node) insert(val string, lastOne bool) (*node, bool) {
	cur := n
	for {
		if cur.val == val {
			if lastOne {
				cur.section = true
			}
			return cur, false
		}

		index, ok := cur.search(val)
		if ok {
			cur = cur.next[index]
			continue
		}

		index = sort.Search(len(cur.next), func(i int) bool {
			return cur.next[i].val >= val
		})

		clone := make([]*node, len(cur.next)+1)
		copy(clone, cur.next[:index])
		clone[index] = &node{val: val, section: lastOne}
		copy(clone[index+1:], cur.next[index:])

		cur.next = clone
		return cur.next[index], true
	}
}

func (n *node) delete(index int) {
	if len(n.next) != 0 {
		n.next = append(n.next[:index], n.next[index+1:]...)
	}
}

func New(fuzzy bool) Trie {
	return &trie{fuzzy: fuzzy, root: new(node)}
}

type trie struct {
	sync.RWMutex
	fuzzy    bool
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
		for _, phrase := range walkNode(cur, delimiter, t.fuzzy) {
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

func walkNode(cur *node, delimiter string, fuzzy bool) (phrases []string) {
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
		switch fuzzy {
		case true:
			if entry.node.leaf() {
				var values []string
				for itor := stack.Front(); itor != nil; itor = itor.Next() {
					values = append(values, itor.Value.(*Entry).node.val)
				}
				values = append(values, entry.node.val)

				phrases = append(phrases, strings.Join(values, delimiter))
			}

		default:
			if entry.node.section && entry.index == 0 {
				var values []string
				for itor := stack.Front(); itor != nil; itor = itor.Next() {
					values = append(values, itor.Value.(*Entry).node.val)
				}
				values = append(values, entry.node.val)

				phrases = append(phrases, strings.Join(values, delimiter))
			}
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

	threshold := len(values) - 1
	for i, val := range values {
		cur, ok = cur.insert(val, i == threshold)
		if ok {
			t.capacity++
		}
	}
}

func (t *trie) HasPrefix(values []string) bool {
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

func (t *trie) Match(values []string) bool {
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

	return cur.section || cur.leaf()
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

	if !t.fuzzy && cur.section && !cur.leaf() { // intermediate node
		t.capacity--
	}
	cur.section = false

	if !cur.leaf() { // not leaf
		return
	}

	last := path[len(path)-1]
	last.node.delete(last.index)
	t.capacity--

	switch t.fuzzy {
	case true:
		for k := len(path) - 2; k >= 0; k-- { // the second last
			if sec := path[k]; last.node.leaf() {
				sec.node.delete(sec.index)
				t.capacity--

				last = sec
				continue
			}

			return
		}

	default:
		for k := len(path) - 2; k >= 0; k-- { // the second last
			if sec := path[k]; last.node.leaf() && !last.node.section {
				sec.node.delete(sec.index)
				last = sec
				continue
			}

			return
		}
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
		phrases = append(phrases, walkNode(next, delimiter, t.fuzzy)...)
	}
	return
}
