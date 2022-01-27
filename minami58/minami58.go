package minami58

import (
	"encoding/binary"
)

const alphabet = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"
const alphabetLen = 58

var index = make(map[byte]uint16, alphabetLen)

func init() {
	for i, c := range alphabet {
		index[byte(c)] = uint16(i)
	}
}

func Encode(raw []byte) []byte {
	if len(raw) == 0 {
		return raw
	}

	var lastOne []byte
	if len(raw)%2 != 0 {
		lastOne = raw[len(raw)-1:]
		raw = raw[:len(raw)-1]
	}

	buf := make([]byte, 0, (len(raw)/2)*3+2)
	for k := 0; k < len(raw); k += 2 {
		num := binary.BigEndian.Uint16(raw[k : k+2])

		x := num / alphabetLen
		a := num % alphabetLen

		b := x / alphabetLen
		c := x % alphabetLen

		buf = append(buf, alphabet[a])
		buf = append(buf, alphabet[b])
		buf = append(buf, alphabet[c])
	}

	if lastOne != nil {
		a := lastOne[0] / alphabetLen
		b := lastOne[0] % alphabetLen

		buf = append(buf, alphabet[a])
		buf = append(buf, alphabet[b])
	}

	return buf
}

func Decode(raw []byte) []byte {
	if len(raw) == 0 {
		return raw
	}

	notDivisible := len(raw)%3 != 0
	if notDivisible && (len(raw)-2)%3 != 0 {
		return raw
	}

	var lastTwo []byte
	if notDivisible {
		lastTwo = raw[len(raw)-2:]
		raw = raw[:len(raw)-2]
	}

	buf := make([]byte, 0, (len(raw)/3)*2+1)
	for k := 0; k < len(raw); k += 3 {
		a := index[raw[k]]
		b := index[raw[k+1]]
		c := index[raw[k+2]]

		x := make([]byte, 2)
		binary.BigEndian.PutUint16(x, (b*alphabetLen+c)*alphabetLen+a)

		buf = append(buf, x...)
	}

	if lastTwo != nil {
		a := uint8(index[lastTwo[0]])
		b := uint8(index[lastTwo[1]])

		x := a*alphabetLen + b
		buf = append(buf, x)
	}

	return buf
}
