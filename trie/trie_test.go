package trie

import (
	"testing"
)

func TestTrie(t *testing.T) {
	trie := New()

	trie.Insert([]string{"a"})
	t.Log(trie.Capacity(), "\n"+trie.String(EmptyDelimiter))

	trie.Insert([]string{"a", "b"})
	t.Log(trie.Capacity(), "\n"+trie.String(EmptyDelimiter))

	trie.Insert([]string{"a", "b", "c"})
	t.Log(trie.Capacity(), "\n"+trie.String(EmptyDelimiter))

	trie.Insert([]string{"b"})
	t.Log(trie.Capacity(), "\n"+trie.String(EmptyDelimiter))

	trie.Insert([]string{"b", "d"})
	t.Log(trie.Capacity(), "\n"+trie.String(EmptyDelimiter))

	trie.Insert([]string{"b", "d", "e"})
	t.Log(trie.Capacity(), "\n"+trie.String(EmptyDelimiter))

	trie.Insert([]string{"b", "d", "f"})
	t.Log(trie.Capacity(), "\n"+trie.String(EmptyDelimiter))

	t.Log(trie.Prompt([]string{"a"}, EmptyDelimiter))

}
