package minami58

import (
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	ts := time.Date(2022, time.July, 7, 27, 47, 57, 0, time.UTC)
	for k := 0; k < 100000000000000000; k++ {
		if ts = ts.Add(time.Minute * 7); ts.Year() >= (epochYear + 15) {
			return
		}

		tx, err := ParseTime(FormatTime(ts))
		if err != nil {
			t.Fatal(err)
		}

		s0 := ts.Format(time.RFC3339)
		s1 := tx.Format(time.RFC3339)
		if s0 != s1 {
			t.Fatal(s0, s1)
		}
	}
}

func BenchmarkTime(b *testing.B) {
	ts0 := time.Now()
	ts1, err := ParseTime(FormatTime(ts0))
	if err != nil {
		b.Fatal(err)
	}

	if ts0.In(time.UTC).Format(time.RFC3339) != ts1.Format(time.RFC3339) {
		b.Fatal("not match")
	}
}
