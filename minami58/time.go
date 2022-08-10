package minami58

import (
	"time"

	"github.com/bluekaki/pkg/errors"
)

const (
	epochYear = 2022 // works for 15 years
)

// FormatTime encode time in minami58, mill-seconds will be ignored.
func FormatTime(ts time.Time) string {
	ts = ts.In(time.UTC)

	val := make([]byte, 6)

	ym := uint8(ts.Year()-epochYear)<<4 | uint8(ts.Month())
	ymOff := uint8(ym / alphabetLen)
	val[0] = alphabet[ym%alphabetLen] // year & month

	val[1] = alphabet[ts.Day()] // day

	val[2] = alphabet[ts.Hour()] // hour

	minuteOff := uint8(ts.Minute() / alphabetLen)
	val[3] = alphabet[ts.Minute()%alphabetLen] // minute

	secondOff := uint8(ts.Second() / alphabetLen)
	val[4] = alphabet[ts.Second()%alphabetLen] // second

	val[5] = alphabet[ymOff<<2|minuteOff<<1|secondOff] // off

	return string(val)
}

// ParseTime parse a minami58 encoded time, return a UTC location time.
func ParseTime(ts string) (time.Time, error) {
	raw := []byte(ts)
	if len(raw) != 6 {
		return time.Time{}, errors.New("ts illegal")
	}

	for _, c := range raw {
		if _, ok := index[c]; !ok {
			return time.Time{}, errors.New("ts illegal")
		}
	}

	offset := uint8(index[raw[5]])

	ym := uint8((offset>>2)*alphabetLen) + uint8(index[raw[0]])
	year := int(ym>>4) + epochYear
	month := time.Month(ym << 4 >> 4)

	day := int(index[raw[1]])
	hour := int(index[raw[2]])

	minute := int((offset<<6>>7)*alphabetLen) + int(index[raw[3]])
	second := int((offset<<7>>7)*alphabetLen) + int(index[raw[4]])

	return time.Date(year, month, day, hour, minute, second, 0, time.UTC), nil
}
