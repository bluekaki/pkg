package sequential

import (
	"bytes"
	"encoding/binary"
)

var (
	emptyIndex = make([]byte, indexLen)
)

type index [3]uint64

func newIndex(offset uint64, dataOffset int64, length int) *index {
	return &index{offset, uint64(dataOffset), uint64(length)}
}

func (m *index) Offset() uint64 {
	return m[0]
}

func (m *index) DataOffset() int64 {
	return int64(m[1])
}

func (m *index) Length() int {
	return int(m[2])
}

func latestBlock(indexRaw [indexSize]byte) (minOffset, maxOffset uint64, dataOffset int64, idx []*index) {
	idx = make([]*index, 0, indexSize/indexLen)
	for k := 0; k < indexSize; k += indexLen {
		raw := indexRaw[k : k+indexLen]
		if bytes.Equal(raw, emptyIndex) {
			return
		}

		if k == 0 {
			minOffset = binary.BigEndian.Uint64(raw[:8])
		}
		maxOffset = binary.BigEndian.Uint64(raw[:8])

		idx = append(idx, newIndex(maxOffset, dataOffset, int(binary.BigEndian.Uint32(raw[8:12]))))
		dataOffset += int64(binary.BigEndian.Uint32(raw[8:12]))
	}
	return
}
