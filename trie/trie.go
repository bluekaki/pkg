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
	String(prefix, delimiter string) string
	Insert(fragments []string)
	HasPrefix(fragments []string) bool
	Match(fragments []string) bool
	Delete(fragments []string)
	Prompt(prefix []string, delimiter string) (phrases []string)
	Marshal() []byte
	Unmarshal(raw []byte) error
}

func SplitByEmpty(phrase string) []string {
	raw := []rune(phrase)
	fragments := make([]string, len(raw))
	for i, char := range raw {
		fragments[i] = string(char)
	}

	return fragments
}

func SplitByDelimiter(phrase, delimiter string) []string {
	raw := strings.Split(strings.TrimLeft(strings.TrimSpace(phrase), delimiter), delimiter)
	fragments := make([]string, len(raw))
	for i, char := range raw {
		fragments[i] = char
	}

	return fragments
}

type node struct {
	key    string
	sector bool
	next   []*node
}

func (n *node) leaf() bool {
	return len(n.next) == 0
}

func (n *node) search(key string) (int, bool) {
	if len(n.next) == 0 {
		return -1, false
	}

	index := sort.Search(len(n.next), func(i int) bool {
		return n.next[i].key >= key
	})

	if index >= len(n.next) || n.next[index].key != key {
		return -1, false
	}

	return index, true
}

func (n *node) insert(key string, lastOne bool) (*node, bool) {
	cur := n
	for {
		if cur.key == key {
			if lastOne {
				cur.sector = true
			}
			return cur, false
		}

		index, ok := cur.search(key)
		if ok {
			cur = cur.next[index]
			continue
		}

		index = sort.Search(len(cur.next), func(i int) bool {
			return cur.next[i].key >= key
		})

		clone := make([]*node, len(cur.next)+1)
		copy(clone, cur.next[:index])
		clone[index] = &node{key: key, sector: lastOne}
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

func (t *trie) String(prefix, delimiter string) string {
	t.RLock()
	defer t.RUnlock()

	buf := bytes.NewBuffer(nil)
	for _, cur := range t.root.next {
		for _, phrase := range walkNode(cur, prefix, delimiter) {
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

func walkNode(cur *node, prefix, delimiter string) (phrases []string) {
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
		if entry.node.sector && entry.index == 0 {
			var fragments []string
			for itor := stack.Front(); itor != nil; itor = itor.Next() {
				fragments = append(fragments, itor.Value.(*Entry).node.key)
			}
			fragments = append(fragments, entry.node.key)

			phrase := strings.Join(fragments, delimiter)
			if prefix != "" {
				phrase = prefix + phrase
			}
			phrases = append(phrases, phrase)
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

func (t *trie) Insert(fragments []string) {
	t.Lock()
	defer t.Unlock()

	cur := t.root
	var ok bool

	threshold := len(fragments) - 1
	for i, key := range fragments {
		cur, ok = cur.insert(key, i == threshold)
		if ok {
			t.capacity++
		}
	}
}

func (t *trie) HasPrefix(fragments []string) bool {
	t.RLock()
	defer t.RUnlock()

	cur := t.root
	for _, key := range fragments {
		index, ok := cur.search(key)
		if !ok {
			return false
		}

		cur = cur.next[index]
	}

	return true
}

func (t *trie) Match(fragments []string) bool {
	t.RLock()
	defer t.RUnlock()

	cur := t.root
	for _, key := range fragments {
		index, ok := cur.search(key)
		if !ok {
			return false
		}

		cur = cur.next[index]
	}

	return cur.sector || cur.leaf()
}

func (t *trie) Delete(fragments []string) {
	t.Lock()
	defer t.Unlock()

	type Entry struct {
		node  *node
		index int
	}
	var path []*Entry

	cur := t.root
	for _, key := range fragments {
		index, ok := cur.search(key)
		if !ok {
			return
		}

		path = append(path, &Entry{node: cur, index: index})
		cur = cur.next[index]
	}

	if cur.sector && !cur.leaf() { // intermediate node
		t.capacity--
	}

	cur.sector = false
	if !cur.leaf() { // not leaf
		return
	}

	last := path[len(path)-1]
	last.node.delete(last.index)
	t.capacity--

	for k := len(path) - 2; k >= 0; k-- { // the second last
		if sec := path[k]; last.node.leaf() && !last.node.sector {
			sec.node.delete(sec.index)
			last = sec
			continue
		}

		return
	}
}

func (t *trie) Prompt(prefix []string, delimiter string) (phrases []string) {
	t.RLock()
	defer t.RUnlock()

	cur := t.root
	for _, key := range prefix {
		index, ok := cur.search(key)
		if !ok {
			return nil
		}

		cur = cur.next[index]
	}

	for _, next := range cur.next {
		phrases = append(phrases, walkNode(next, "", delimiter)...)
	}
	return
}

func (t *trie) Marshal() []byte {
	// TODO
	return nil
}

func (t *trie) Unmarshal([]byte) error {
	// TODO
	return nil
}
