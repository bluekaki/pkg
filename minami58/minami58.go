package minami58

import (
	"crypto/sha256"
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

// Shorten add some salt if in duplicated
func Shorten(url string) string {
	digest := sha256.Sum256([]byte(url))

	link := []byte{
		lookup([5]byte{digest[0], digest[1], digest[2], digest[3], digest[4] >> 4}),
		lookup([5]byte{digest[4] << 4 >> 4, digest[5], digest[6], digest[7], digest[8]}),
		lookup([5]byte{digest[9], digest[10], digest[11], digest[12], digest[13] >> 4}),
		lookup([5]byte{digest[13] << 4 >> 4, digest[14], digest[15], digest[16], digest[17]}),
		lookup([5]byte{digest[18], digest[19], digest[20], digest[21], digest[22] >> 4}),
		lookup([5]byte{digest[22] << 4 >> 4, digest[23], digest[24], digest[25], digest[26]}),
		lookup([5]byte{digest[27], digest[28], digest[29], digest[30], digest[31] >> 4}),
	}
	return string(link)
}

func lookup(raw [5]byte) byte {
	offset := uint8(0)
	for _, x := range []uint8{
		raw[0] >> 2,
		raw[1] >> 2,
		raw[2] >> 2,
		raw[0]<<6>>2 | raw[1]<<6>>4 | raw[2]<<6>>6,
		raw[3] >> 2,
		raw[3]<<6>>2 | raw[4],
	} {
		if offset += x % alphabetLen; offset >= alphabetLen {
			offset -= alphabetLen
		}
	}

	return alphabet[offset]
}
