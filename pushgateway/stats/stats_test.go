package stats

import (
	"testing"
	"time"

	"github.com/bluekaki/pkg/httpclient"
)

func TestParseMetrics(t *testing.T) {
	body, _, _, err := httpclient.Get("http://127.0.0.1:9091/api/v1/metrics", nil,
		httpclient.WithRetryTimes(0),
		httpclient.WithTTL(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	info, err := ParseMetrics(body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(info))
}

func TestDeleteAll(t *testing.T) {
	if err := DeleteAll("127.0.0.1:9091"); err != nil {
		t.Fatal(err)
	}
}
