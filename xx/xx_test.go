package xx

import (
	"testing"
)

func TestParse(t *testing.T) {
	t.Log(Parse("/a/b/c/"))

	t.Log(Parse(" ///a/b/c/// "))

	t.Log(Parse("/a//b/c/"))

	t.Log(Parse("/**//b/c/"))

	t.Log(Parse("/a/**/b/c/"))

	t.Log(Parse("/a/*/b/c/**"))

	t.Log(Parse("/a/*/b/c/**///"))
}

func TestInsert(t *testing.T) {
	tree := New()
	insert := func(pattern string) {
		if err := tree.Insert(pattern); err != nil {
			t.Error(err)
		}
	}

	insert("/a/b/b/b/b/c")
	insert("/a/b/c/d")
	insert("/a/b/c/c/c/c/d/e/**")
	insert("/a/b/c/c/c/c/d/e/*/*/*/**")

	t.Log("\n" + tree.String())

}
