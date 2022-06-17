package buffer

import (
	"io"
	"sync"
)

const defaultCapacity = 4 << 10

func NewFixedBuffer(capacity int) FixedBuffer {
	if capacity <= 0 {
		capacity = defaultCapacity
	}

	return &buffer{
		capacity: capacity,
		raw:      make([]byte, capacity),
	}
}

type FixedBuffer interface {
	io.Writer
	Bytes() []byte
}

type buffer struct {
	sync.RWMutex
	capacity int
	raw      []byte
	index    int
}

func (b *buffer) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}

	b.Lock()
	defer b.Unlock()

	if len(p) >= b.capacity {
		copy(b.raw, p)
		b.index = b.capacity
		return b.capacity, nil
	}

	payloadLen := len(p)
	if (b.index + payloadLen) <= b.capacity {
		copy(b.raw[b.index:], p)
		b.index += payloadLen

		return payloadLen, nil
	}

	remainder := b.index + payloadLen - b.capacity
	copy(b.raw, b.raw[remainder:])
	copy(b.raw[b.capacity-payloadLen:], p)
	b.index = b.capacity

	return payloadLen, nil
}

func (b *buffer) Read(p []byte) (n int, err error) {

	return
}

func (b *buffer) Bytes() []byte {
	b.RLock()
	defer b.RUnlock()

	return b.raw[:b.index]
}
