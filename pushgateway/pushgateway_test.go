package pushgateway

import (
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	porter := NewPorter("http://127.0.0.1:8080/metrics", "http://127.0.0.1:9091", WithLabel("project", "dummy"))

	go func() {
		for err := range porter.Errors() {
			panic(err)
		}
	}()

	time.Sleep(time.Second * 3)
	if err := porter.Close(); err != nil {
		t.Fatal(err)
	}
}
