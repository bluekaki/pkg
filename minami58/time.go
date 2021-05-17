package minami58

import (
	"time"

	"github.com/bluekaki/pkg/errors"
)

const (
	epochYear = 2021
)

// FormatTime encode time in minami58, mill-seconds will be ignored.
func FormatTime(ts time.Time) string {
	ts = ts.In(time.UTC)

	val := make([]byte, 6)

	ym := uint8(ts.Year()-epochYear)<<4 | uint8(ts.Month())
	ymOff := uint8(ym / alphabets)
	val[0] = alphabet[ym%alphabets] // year & month

	val[1] = alphabet[ts.Day()] // day

	hourOff := uint8(ts.Hour() / alphabets)
	val[2] = alphabet[ts.Hour()%alphabets] // hour

	minuteOff := uint8(ts.Minute() / alphabets)
	val[3] = alphabet[ts.Minute()%alphabets] // minute

	secondOff := uint8(ts.Second() / alphabets)
	val[4] = alphabet[ts.Second()%alphabets] // second

	val[5] = alphabet[ymOff<<3|hourOff<<2|minuteOff<<1|secondOff] // off

	return string(val)
}

// ParseTime parse a minami58 encoded time, return a UTC location time.
func ParseTime(ts string) (time.Time, error) {
	raw := []byte(ts)
	if len(raw) != 6 {
		return time.Time{}, errors.New("ts illegal")
	}

	for _, c := range raw {
		if int(c) > len(reversedAlphabet) || reversedAlphabet[c] == -1 {
			return time.Time{}, errors.New("ts illegal")
		}
	}

	offset := uint8(reversedAlphabet[raw[5]])

	ym := uint8((offset>>3)*alphabets) + uint8(reversedAlphabet[raw[0]])
	year := int(ym>>4) + epochYear
	month := time.Month(ym << 4 >> 4)

	day := reversedAlphabet[raw[1]]

	hour := int((offset<<5>>7)*alphabets) + reversedAlphabet[raw[2]]
	minute := int((offset<<6>>7)*alphabets) + reversedAlphabet[raw[3]]
	second := int((offset<<7>>7)*alphabets) + reversedAlphabet[raw[4]]

	return time.Date(year, month, day, hour, minute, second, 0, time.UTC), nil
}
