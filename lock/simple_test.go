package lock

import (
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
)

func TestSimple(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:               "xx.com:6379",
		MaxRetries:         3,
		DialTimeout:        time.Second * 2,
		ReadTimeout:        time.Second,
		WriteTimeout:       time.Second,
		PoolSize:           50,
		MinIdleConns:       5,
		MaxConnAge:         time.Minute,
		PoolTimeout:        time.Minute,
		IdleTimeout:        time.Second * 30,
		IdleCheckFrequency: time.Second,
	})
	if err := client.Ping().Err(); err != nil {
		t.Fatal(err)
	}

	counter := 0

	index := 0
	handler := func() error {
		if val := client.PTTL("xxxxxx").Val(); val > 0 {
			t.Log(index, val)
			index++
			return nil
		}

		if err := client.Set("xxxxxx", "", time.Second*10).Err(); err != nil {
			t.Fatal(err)
		}

		t.Log(counter)
		counter++
		return nil
	}

	wg := new(sync.WaitGroup)
	wg.Add(10)

	for k := 0; k < 10; k++ {
		go func() {
			defer wg.Done()

			if err := Simple("xxx", handler, WithRedisClient(client)); err != nil {
				t.Fatal(err)
			}
		}()
	}

	wg.Wait()
}
