package sequential

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

var (
	emptyIndex = make([]byte, indexLen)
)

func reversedIndex(indexRaw [indexSize]byte) func() (minOffset, maxOffset uint64, digest []byte, index *index, err error) {
	type Entry struct {
		minOffset  uint64
		maxOffset  uint64
		dataOffset int64
		digest     []byte
		index      *index
	}

	var last *Entry
	var entries []*Entry

	for k := 0; k < indexSize; k += indexLen {
		entry := new(Entry)
		if k == 0 {
			entry.index = newIndex()

		} else {
			last = entries[len(entries)-1]

			entry.minOffset = last.minOffset
			entry.dataOffset = last.dataOffset
			entry.index = last.index.Clone()
		}

		raw := indexRaw[k : k+indexLen]
		if bytes.Equal(raw, emptyIndex) {
			break
		}

		if k == 0 {
			entry.minOffset = binary.BigEndian.Uint64(raw[:8])
		}
		entry.maxOffset = binary.BigEndian.Uint64(raw[:8])

		entry.index.AppendEntry(newEntry(entry.maxOffset, entry.dataOffset, int(binary.BigEndian.Uint32(raw[8:12]))))
		entry.dataOffset += int64(binary.BigEndian.Uint32(raw[8:12]))
		entry.digest = raw[12:]

		entries = append(entries, entry)
	}

	i, j := 0, len(entries)-1
	for i <= j {
		entries[i], entries[j] = entries[j], entries[i]

		i++
		j--
	}

	return func() (minOffset, maxOffset uint64, digest []byte, index *index, err error) {
		if len(entries) == 0 {
			err = io.EOF
			return
		}

		entry := entries[0]
		minOffset = entry.minOffset
		maxOffset = entry.maxOffset
		digest = entry.digest
		index = entry.index

		entries = entries[1:]
		return
	}
}

func encodeCapacity(minOffset, maxOffset uint64) []byte {
	return []byte(fmt.Sprintf("%04d", maxOffset-minOffset+1))
}
