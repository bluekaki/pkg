package id

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/bluekaki/pkg/errors"
)

// ErrSequenceOverflow oveflow error
var ErrSequenceOverflow = fmt.Errorf("sequence over the max limition of %d", MaxSequence)

const (
	epochYear = 2021
	endYear   = epochYear + 15

	// MaxSequence maximum allowed value
	MaxSequence uint32 = math.MaxUint32 >> 4 // 268435455
)

// Prefix two visible ASCII
type Prefix [2]byte

func (p Prefix) String() string {
	return string(p[:])
}

// Sequence number generator
type Sequence interface {
	Next() (uint32, error)
}

// Generator genarate a 18len decimal id with 2len prefix
type Generator interface {
	New(prefix Prefix) (string, error)
	Parse(id string) (Prefix, time.Time, error)
}

var _ Generator = (*generator)(nil)

type generator struct {
	sequence Sequence
}

// NewGenerator create a new generator
func NewGenerator(sequence Sequence) Generator {
	if sequence == nil {
		panic("sequence required")
	}

	return &generator{sequence: sequence}
}

func (g *generator) New(prefix Prefix) (string, error) {
	sequence, err := g.sequence.Next()
	if err != nil {
		return "", err
	}

	if sequence > MaxSequence {
		return "", ErrSequenceOverflow
	}

	ts := time.Now().In(time.UTC)
	val := uint64(1)<<58 |
		uint64(ts.Year()-epochYear)<<54 |
		uint64(ts.Month())<<50 |
		uint64(ts.Day())<<45 |
		uint64(ts.Hour())<<40 |
		uint64(ts.Minute())<<34 |
		uint64(ts.Second())<<28 |
		uint64(sequence)

	raw := []byte(strconv.FormatUint(val, 2))

	shuffle := make([]byte, 0, 59)
	shuffle = append(shuffle, raw[0])
	for k := 1; k < 30; k++ {
		shuffle = append(shuffle, raw[k])
		shuffle = append(shuffle, raw[59-k])
	}

	val, _ = strconv.ParseUint(string(shuffle), 2, 64)
	return prefix.String() + strconv.FormatUint(val, 10), nil
}

func (g *generator) Parse(id string) (Prefix, time.Time, error) {
	if len(id) != 20 {
		return [2]byte{'X', 'X'}, time.Time{}, errors.New("id length must be 20")
	}

	val, err := strconv.ParseUint(id[2:], 10, 64)
	if err != nil {
		return [2]byte{'X', 'X'}, time.Time{}, errors.Wrapf(err, "parse id [%s] to uint err", id)
	}

	const max = math.MaxUint64 >> 5
	if val > max {
		return [2]byte{'X', 'X'}, time.Time{}, errors.New("id illegal")
	}

	shuffle := []byte(strconv.FormatUint(val, 2))
	if len(shuffle) != 59 {
		return [2]byte{'X', 'X'}, time.Time{}, errors.New("id illegal")
	}

	raw := make([]byte, 59)
	raw[0] = shuffle[0]
	for k, index := 1, 1; k < 30; k++ {
		raw[k] = shuffle[index]
		raw[59-k] = shuffle[index+1]
		index += 2
	}

	val, _ = strconv.ParseUint(string(raw), 2, 64)

	var ts time.Time
	{
		year := val<<6>>60 + epochYear
		month := val << 10 >> 60
		day := val << 14 >> 59
		hour := val << 19 >> 59
		minute := val << 24 >> 58
		second := val << 30 >> 58

		ts, err = time.ParseInLocation("2006-01-02T15:04:05", fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d", year, month, day, hour, minute, second), time.UTC)
		if err != nil {
			return [2]byte{'X', 'X'}, time.Time{}, errors.Wrap(err, "id illegal")
		}
	}

	return [2]byte{id[0], id[1]}, ts.In(time.Local), nil
}
