package rate

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

var instance Limiter

func TestMain(m *testing.M) {
	conf := new(Config)
	conf.Identifier = "dummy-test"
	conf.TickerInterval = time.Millisecond * 200
	conf.Limit.Upper = 1000
	conf.Limit.Lower = 100
	conf.Log.Enable = false

	redis, err := NewSingleRedis("127.0.0.1:6379", "", 0, nil)
	if err != nil {
		panic(err)
	}

	if instance, err = NewLimiter(conf, redis); err != nil {
		panic(err)
	}

	m.Run()
	instance.Close()
}

func TestRate(t *testing.T) {
	summary := uint64(0)

	for k := 0; k < 2000; k++ {
		go func() {
			for {
				atomic.AddUint64(&summary, 1)
				instance.Allow()
			}
		}()
	}

	do := func(ttl time.Duration) {
		lastSummary := atomic.LoadUint64(&summary)
		ctx, cancel := context.WithTimeout(context.Background(), ttl)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				nowSummary := atomic.LoadUint64(&summary)
				t.Log("rate", nowSummary-lastSummary)
				lastSummary = nowSummary
			}
		}
	}

	do(time.Second * 30)
	t.Log("------update rate------")

	instance.UpdateRate(100, 10)
	do(time.Second * 30)
}
