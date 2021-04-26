package minami58

import (
	"github.com/byepichi/pkg/errors"
)

// Coefficient of expansion 1.5

const (
	step   = 16 // 4bits
	cycles = 29 // 29 × 16 ÷ 58 = 8

	alphabets = 58 // 58 alphabets
)

var (
	alphabet         = []byte("NWbqy6aVDPJQv7MmSKGkEYhTFciH1gZ8endpXsCUL93ofRAB5zjx4tur2w")
	reversedAlphabet ['z' + 1]int

	dict    [cycles][]byte        // (1/2)bytes => 1byte
	mapping [cycles]['z' + 1]byte // 1byte => 2bytes
)

func init() {
	for i := range reversedAlphabet {
		reversedAlphabet[i] = -1
	}

	for i, c := range alphabet {
		reversedAlphabet[c] = i
	}
}

func init() {
	nexter := builder()
	for k := 0; k < cycles; k++ {
		source := nexter()
		dict[k] = source

		for i, c := range source {
			mapping[k][c] = byte(i)
		}
	}
}

type nexter func() []byte

func builder() nexter {
	cursor := 0
	return func() (raw []byte) {
		next := cursor + step

		if offset := next - len(alphabet); offset > 0 {
			raw = append(alphabet[cursor:], alphabet[:offset]...)
			cursor = offset

		} else {
			raw = alphabet[cursor:next]
			cursor = next
		}

		return
	}
}

// Encode something
func Encode(payload []byte) []byte {
	if len(payload) == 0 {
		return nil
	}

	raw := make([]byte, len(payload)*2)
	for i, v := range payload {
		x := i * 2
		raw[x] = dict[x%cycles][v>>4]

		y := i*2 + 1
		raw[y] = dict[y%cycles][v<<4>>4]
	}

	return raw
}

// Decode something
func Decode(raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	if len(raw)%2 != 0 {
		return nil, errors.New("raw should be in pairs")
	}

	payload := make([]byte, len(raw)/2)

	size := len(raw)
	for i := 0; i < size; i += 2 {
		x := i % cycles
		y := (i + 1) % cycles

		payload[i/2] = mapping[x][raw[i]]<<4 | mapping[y][raw[i+1]]
	}

	return payload, nil
}
