package pbutil

import (
	"testing"
	"time"
)

func TestTimestamp(t *testing.T) {
	ts0 := time.Now()

	timestamp, err := NewTimestamp(&ts0)
	if err != nil {
		t.Fatal(err)
	}

	ts1, err := ParseTimestamp(timestamp)
	if err != nil {
		t.Fatal(err)
	}

	if ts0.Format(time.RFC3339Nano) != ts1.Format(time.RFC3339Nano) {
		t.Fatal("timestamp not match")
	}
}

func TestDuration(t *testing.T) {
	d0 := time.Second * 4

	duration := NewDuration(d0)

	d1, err := ParseDuration(duration)
	if err != nil {
		t.Fatal(err)
	}

	if d0 != d1 {
		t.Fatal("duration not match")
	}
}
