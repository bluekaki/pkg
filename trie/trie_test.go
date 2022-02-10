package trie

import (
	"testing"
)

func TestTrie(t *testing.T) {
	tree := New(true)

	tree.Insert(SplitByEmpty("a"))
	t.Log(tree.Capacity(), len(tree.Prompt(SplitByEmpty(""), EmptyDelimiter)), "\n"+tree.String(EmptyDelimiter))

	tree.Insert(SplitByEmpty("ab"))
	t.Log(tree.Capacity(), len(tree.Prompt(SplitByEmpty(""), EmptyDelimiter)), "\n"+tree.String(EmptyDelimiter))

	tree.Insert(SplitByEmpty("abc"))
	t.Log(tree.Capacity(), len(tree.Prompt(SplitByEmpty(""), EmptyDelimiter)), "\n"+tree.String(EmptyDelimiter))

	tree.Insert(SplitByEmpty("b"))
	t.Log(tree.Capacity(), len(tree.Prompt(SplitByEmpty(""), EmptyDelimiter)), "\n"+tree.String(EmptyDelimiter))

	tree.Insert(SplitByEmpty("bd"))
	t.Log(tree.Capacity(), len(tree.Prompt(SplitByEmpty(""), EmptyDelimiter)), "\n"+tree.String(EmptyDelimiter))

	tree.Insert(SplitByEmpty("bde"))
	t.Log(tree.Capacity(), len(tree.Prompt(SplitByEmpty(""), EmptyDelimiter)), "\n"+tree.String(EmptyDelimiter))

	tree.Insert(SplitByEmpty("bdf"))
	t.Log(tree.Capacity(), len(tree.Prompt(SplitByEmpty(""), EmptyDelimiter)), "\n"+tree.String(EmptyDelimiter))

	t.Log(tree.Prompt(SplitByEmpty("a"), EmptyDelimiter))

	tree.Delete(SplitByEmpty("b"))
	t.Log(tree.Capacity(), len(tree.Prompt(SplitByEmpty(""), EmptyDelimiter)), "\n"+tree.String(EmptyDelimiter))

	tree.Delete(SplitByEmpty("bd"))
	t.Log(tree.Capacity(), len(tree.Prompt(SplitByEmpty(""), EmptyDelimiter)), "\n"+tree.String(EmptyDelimiter))

	tree.Delete(SplitByEmpty("bde"))
	t.Log(tree.Capacity(), len(tree.Prompt(SplitByEmpty(""), EmptyDelimiter)), "\n"+tree.String(EmptyDelimiter))

	tree.Delete(SplitByEmpty("bdf"))
	t.Log(tree.Capacity(), len(tree.Prompt(SplitByEmpty(""), EmptyDelimiter)), "\n"+tree.String(EmptyDelimiter))

	tree.Delete(SplitByEmpty("abc"))
	t.Log(tree.Capacity(), len(tree.Prompt(SplitByEmpty(""), EmptyDelimiter)), "\n"+tree.String(EmptyDelimiter))

	tree.Delete(SplitByEmpty("ab"))
	t.Log(tree.Capacity(), len(tree.Prompt(SplitByEmpty(""), EmptyDelimiter)), "\n"+tree.String(EmptyDelimiter))

	tree.Delete(SplitByEmpty("a"))
	t.Log(tree.Capacity(), len(tree.Prompt(SplitByEmpty(""), EmptyDelimiter)), "\n"+tree.String(EmptyDelimiter))
}
