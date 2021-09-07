package sequential

import (
	"bytes"
	"encoding/binary"
)

var (
	emptyIndex = make([]byte, indexLen)
)

func latestBlock(indexRaw []byte) (blocks int, minOffset, maxOffset uint64, dataOffset int64) {
	for k := 0; k < indexSize; k += indexLen {
		index := indexRaw[k : k+indexLen]
		if bytes.Equal(index, emptyIndex) {
			return
		}

		blocks++
		dataOffset += int64(binary.BigEndian.Uint32(index[8:12]))

		if k == 0 {
			minOffset = binary.BigEndian.Uint64(index[:8])
		}
		maxOffset = binary.BigEndian.Uint64(index[:8])
	}
	return
}
