package timeutil

import (
	"testing"
)

func TestTimezone(t *testing.T) {
	t.Log("CST", NowInCST())
}
