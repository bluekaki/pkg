package pushgateway

import (
	"testing"
)

func TestRun(t *testing.T) {
	for err := range Run("http://127.0.0.1:8080/metrics", "http://127.0.0.1:9091", WithLabel("project", "dummy")) {
		t.Fatal(err)
	}
}
