package sequential

import (
	"fmt"
	"os"
	"sort"

	"github.com/bluekaki/pkg/errors"
)

type block struct {
	fileIndex uint64
	file      *os.File
	minOffset uint64
	maxOffset uint64
	index     *index
}

type blocks struct {
	slice []*block
}

func newBlocks() *blocks {
	return &blocks{slice: make([]*block, 0, 10)}
}

func (b *blocks) Close() {
	for _, block := range b.slice {
		block.file.Close()
	}
}

func (b *blocks) String() {
	for i, block := range b.slice {
		fmt.Println(i, "fileIndex:", block.fileIndex, "minOffset:", block.minOffset, "maxOffset:", block.maxOffset, "index:", block.index.String())
	}
}

func (b *blocks) AppendAndSort(block *block) {
	b.slice = append(b.slice, block)
	sort.Slice(b.slice, func(i, j int) bool {
		return b.slice[i].fileIndex < b.slice[j].fileIndex
	})
}

func (b *blocks) Append(block *block) {
	b.slice = append(b.slice, block)
}

func (b *blocks) Last() *block {
	if len(b.slice) == 0 {
		return nil
	}

	return b.slice[len(b.slice)-1]
}

func (b *blocks) UpdateLast(entry *entry, minOffset, maxOffset uint64) {
	if last := b.Last(); last != nil {
		last.index.AppendEntry(entry)
		last.minOffset = minOffset
		last.maxOffset = maxOffset
	}
}

func (b *blocks) Get(offset uint64) ([]byte, error) {
	last := b.Last()
	if last == nil {
		return nil, ErrNotfound
	}

	slice := b.slice
	if last.maxOffset == 0 {
		slice = slice[:len(slice)-1]
	}

	index := sort.Search(len(slice), func(i int) bool {
		return offset <= slice[i].maxOffset
	})

	if index == -1 || index >= len(slice) || slice[index].minOffset > offset {
		return nil, ErrNotfound
	}

	block := slice[index]

	entry := block.index.Get(offset)
	if entry == nil {
		panic(fmt.Sprintf("not found offset %d in a sure index of file %s", offset, block.file.Name()))
	}

	raw := make([]byte, entry.Length())
	if _, err := block.file.ReadAt(raw, dataOffset+entry.DataOffset()); err != nil {
		return nil, errors.Wrapf(err, "read offset %d in file %s err", offset, block.file.Name())
	}

	return raw, nil
}
