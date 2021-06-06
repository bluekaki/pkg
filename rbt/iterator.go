package rbt

// Iterator holding the iterator's state
type Iterator struct {
	tree     *rbTree
	node     *node
	position position
}

type position byte

const (
	begin, between, end position = 0, 1, 2
)

// Iterator returns a stateful iterator whose elements are key/value pairs.
func (t *rbTree) Iterator() Iterator {
	return Iterator{tree: t, node: nil, position: begin}
}

// Next moves the iterator to the next element and returns true if there was a next element in the container.
// If Next() returns true, then next element's key and value can be retrieved by Key() and Value().
// If Next() was called for the first time, then it will point the iterator to the first element if it exists.
// Modifies the state of the iterator.
func (iterator *Iterator) Next() bool {
	if iterator.position == end {
		goto end
	}
	if iterator.position == begin {
		left := iterator.tree.minimum(iterator.tree.root)
		if left == nil {
			goto end
		}
		iterator.node = left
		goto between
	}
	if iterator.node.R != nil {
		iterator.node = iterator.node.R
		for iterator.node.L != nil {
			iterator.node = iterator.node.L
		}
		goto between
	}
	if iterator.node.P != nil {
		node := iterator.node
		for iterator.node.P != nil {
			iterator.node = iterator.node.P
			if node.val <= iterator.node.val {
				goto between
			}
		}
	}

end:
	iterator.node = nil
	iterator.position = end
	return false

between:
	iterator.position = between
	return true
}

// Prev moves the iterator to the previous element and returns true if there was a previous element in the container.
// If Prev() returns true, then previous element's key and value can be retrieved by Key() and Value().
// Modifies the state of the iterator.
func (iterator *Iterator) Prev() bool {
	if iterator.position == begin {
		goto begin
	}
	if iterator.position == end {
		right := iterator.tree.maximum(iterator.tree.root)
		if right == nil {
			goto begin
		}
		iterator.node = right
		goto between
	}
	if iterator.node.L != nil {
		iterator.node = iterator.node.L
		for iterator.node.R != nil {
			iterator.node = iterator.node.R
		}
		goto between
	}
	if iterator.node.P != nil {
		node := iterator.node
		for iterator.node.P != nil {
			iterator.node = iterator.node.P
			if node.val >= iterator.node.val {
				goto between
			}
		}
	}

begin:
	iterator.node = nil
	iterator.position = begin
	return false

between:
	iterator.position = between
	return true
}

// Value returns the current element's value.
// Does not modify the state of the iterator.
func (iterator *Iterator) Value() int {
	return iterator.node.val
}

// Begin resets the iterator to its initial state (one-before-first)
// Call Next() to fetch the first element if any.
func (iterator *Iterator) Begin() {
	iterator.node = nil
	iterator.position = begin
}

// End moves the iterator past the last element (one-past-the-end).
// Call Prev() to fetch the last element if any.
func (iterator *Iterator) End() {
	iterator.node = nil
	iterator.position = end
}

// First moves the iterator to the first element and returns true if there was a first element in the container.
// If First() returns true, then first element's key and value can be retrieved by Key() and Value().
// Modifies the state of the iterator
func (iterator *Iterator) First() bool {
	iterator.Begin()
	return iterator.Next()
}

// Last moves the iterator to the last element and returns true if there was a last element in the container.
// If Last() returns true, then last element's key and value can be retrieved by Key() and Value().
// Modifies the state of the iterator.
func (iterator *Iterator) Last() bool {
	iterator.End()
	return iterator.Prev()
}
