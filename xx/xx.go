package xx

import (
	"bytes"
	"container/list"
	stderr "errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

const (
	empty     = ""
	fuzzy     = "*"
	omitted   = "**"
	delimiter = "/"
)

func Parse(pattern string) ([]string, error) {
	const format = "{a-Z}+/{*}+/{**}"

	if pattern = strings.Trim(strings.TrimSpace(pattern), delimiter); pattern == "" {
		return nil, fmt.Errorf("pattern illegal, should in format of %s", format)
	}

	fragments := strings.Split(pattern, delimiter)
	if len(fragments) < 2 {
		return nil, fmt.Errorf("pattern illegal, should in format of %s", format)
	}

	for i := range fragments {
		fragments[i] = strings.TrimSpace(fragments[i])
	}

	// likes **
	if fragments[0] == omitted {
		return nil, stderr.New("illegal omitted")
	}

	for k := 0; k < len(fragments); k++ {
		if fragments[k] == empty {
			return nil, stderr.New("pattern contains empty path")
		}

		if fragments[k] == omitted && k+1 != len(fragments) {
			return nil, stderr.New("pattern contains illegal omitted path")
		}
	}

	return fragments, nil
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

func (n *node) insert(cursor int, fragments []string, paths *[]string) (*node, bool, error) {
	var ok bool
	if cursor > 0 {
		ok = fragments[cursor-1] != fragments[cursor]
	}

	lastOne := cursor == len(fragments)-1

	cur := n
	for {
		if ok && cur.key == fragments[cursor] {
			if lastOne {
				cur.sector = true
			}
			return cur, false, nil
		}
		ok = true

		if cur.key != "" {
			*paths = append(*paths, cur.key)
		}

		index, ok := cur.search(fragments[cursor])
		if ok {
			cur = cur.next[index]
			continue
		}

		if fragments[cursor] == omitted {
			if fragments := n.fuzzy(); len(fragments) > 0 {
				*paths = append(*paths, fragments[:]...)
				return nil, false, stderr.New("pattern conflict")
			}

		} else if _, ok := cur.search(omitted); ok &&
			fragments[cursor] == fuzzy &&
			countFuzzy(fragments[cursor:]) == len(fragments[cursor:]) {
			*paths = append(*paths, omitted)
			return nil, false, stderr.New("pattern conflict")
		}

		index = sort.Search(len(cur.next), func(i int) bool {
			return cur.next[i].key >= fragments[cursor]
		})

		clone := make([]*node, len(cur.next)+1)
		copy(clone, cur.next[:index])
		clone[index] = &node{key: fragments[cursor], sector: lastOne}
		copy(clone[index+1:], cur.next[index:])

		cur.next = clone
		return cur.next[index], true, nil
	}
}

func (n *node) delete(index int) {
	if len(n.next) != 0 {
		n.next = append(n.next[:index], n.next[index+1:]...)
	}
}

func countFuzzy(fragments []string) (summary int) {
	for _, fragment := range fragments {
		if fragment == fuzzy || fragment == omitted {
			summary++
		}
	}
	return
}

func (n *node) fuzzy() []string {
	for _, next := range n.next {
		for _, fragments := range walkNode(next) {
			if countFuzzy(fragments) == len(fragments) {
				return fragments
			}
		}
	}
	return nil
}

func New() *trie {
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
func (t *trie) String() string {
	t.RLock()
	defer t.RUnlock()

	buf := bytes.NewBuffer(nil)
	for _, cur := range t.root.next {
		for _, fragments := range walkNode(cur) {
			buf.WriteString(delimiter)
			buf.WriteString(strings.Join(fragments, delimiter))
			buf.WriteString("\n")
		}
	}

	msg := buf.String()
	if len(msg) > 0 {
		msg = msg[:len(msg)-1] // trim \n
	}

	return msg
}

func walkNode(cur *node) (phrases [][]string) {
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

			phrases = append(phrases, fragments)
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

func (t *trie) Insert(pattern string) error {
	fragments, err := Parse(pattern)
	if err != nil {
		return err
	}

	t.Lock()
	defer t.Unlock()

	cur := t.root
	var ok bool
	var paths []string

	for i := range fragments {
		cur, ok, err = cur.insert(i, fragments, &paths)
		if err != nil {
			return fmt.Errorf("pattern conflict with /%s and /%s", strings.Join(paths, delimiter), strings.Join(fragments, delimiter))
		}
		if ok {
			t.capacity++
		}
	}

	return nil
}
