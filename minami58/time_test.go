package minami58

import (
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	ts := time.Now()
	t.Log(ts)

	val := FormatTime(ts)
	t.Log(val)

	ts, err := ParseTime(val)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ts)
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
