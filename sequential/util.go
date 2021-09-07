package sequential

import (
	"bytes"
	"encoding/binary"
)

var (
	emptyIndex = make([]byte, indexLen)
)

func latestBlock(indexRaw [indexSize]byte) (minOffset, maxOffset uint64, dataOffset int64, index *index) {
	index = newIndex()
	for k := 0; k < indexSize; k += indexLen {
		raw := indexRaw[k : k+indexLen]
		if bytes.Equal(raw, emptyIndex) {
			return
		}

		if k == 0 {
			minOffset = binary.BigEndian.Uint64(raw[:8])
		}
		maxOffset = binary.BigEndian.Uint64(raw[:8])

		index.AppendEntry(newEntry(maxOffset, dataOffset, int(binary.BigEndian.Uint32(raw[8:12]))))
		dataOffset += int64(binary.BigEndian.Uint32(raw[8:12]))
	}
	return
}
