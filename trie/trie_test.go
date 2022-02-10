package trie

import (
	"testing"
)

func TestTrie(t *testing.T) {
	trie := new(Trie)

	trie.Insert("a")
	trie.Insert("ab")
	trie.Insert("abc")
	trie.Insert("b")
	trie.Insert("bd")
	trie.Insert("bde")
	trie.Insert("bdf")

	t.Log("\n" + trie.String())
}
