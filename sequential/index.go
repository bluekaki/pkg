package sequential

import (
	"fmt"
	"sort"
)

type entry [3]uint64

func newEntry(offset uint64, dataOffset int64, length int) *entry {
	return &entry{offset, uint64(dataOffset), uint64(length)}
}

func (e *entry) Offset() uint64 {
	return e[0]
}

func (e *entry) DataOffset() int64 {
	return int64(e[1])
}

func (e *entry) Length() int {
	return int(e[2])
}

type index []*entry

func newIndex() *index {
	idx := make(index, 0, indexSize/indexLen)
	return &idx
}

func (i *index) Clone() *index {
	index := newIndex()
	for _, entry := range *i {
		index.AppendEntry(entry)
	}

	return index
}

func (i *index) AppendEntry(entry *entry) {
	*i = append(*i, entry)
}

func (i *index) Last() *entry {
	if len(*i) == 0 {
		return nil
	}

	return (*i)[len(*i)-1]
}

func (i *index) String() string {
	var slice [][3]uint64
	for _, entry := range *i {
		slice = append(slice, [3]uint64(*entry))
	}

	return fmt.Sprintf("%v", slice)
}

func (i *index) Get(offset uint64) *entry {
	index := sort.Search(len(*i), func(j int) bool {
		return (*i)[j].Offset() >= offset
	})
	if index == -1 || index >= len(*i) || (*i)[index].Offset() != offset {
		return nil
	}

	return (*i)[index]
}
