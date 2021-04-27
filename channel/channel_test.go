package channel

import (
	"context"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTypeAssert(t *testing.T) {
	assert := assert.New(t)

	consumer := func(value interface{}) {
	}

	channel, err := NewChannel(reflect.TypeOf(int(0)), 10, consumer)
	assert.Nil(err)

	assert.Nil(channel.Append(1))
	assert.NotNil(channel.Append("1"))
	assert.Equal(channel.T(), "int")

}

func TestChannel(t *testing.T) {
	assert := assert.New(t)

	summary := 0
	consumer := func(value interface{}) {
		summary += value.(int)
	}

	channel, err := NewChannel(reflect.TypeOf(int(0)), 10, consumer)
	assert.Nil(err)

	max := 100
	for k := 0; k < max; k++ {
		err = channel.Append(1)
		assert.Nil(err)
	}

	channel.Close()
	assert.Equal(max, summary)
}

func TestChannelConcurrent(t *testing.T) {
	assert := assert.New(t)

	var summary uint64
	consumer := func(value interface{}) {
		atomic.AddUint64(&summary, value.(uint64))
	}

	channel, err := NewChannel(reflect.TypeOf(uint64(0)), 65535, consumer, WithWorker(50))
	assert.Nil(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	var max uint64
	wg := new(sync.WaitGroup)

	for k := 0; k < 50; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					err = channel.Append(uint64(1))
					assert.Nil(err)
					atomic.AddUint64(&max, 1)
				}
			}
		}()
	}

	<-ctx.Done()
	channel.Close()

	wg.Wait()
	t.Log(max, summary)
	assert.Equal(max, summary)
}

func BenchmarkTypeOf(b *testing.B) {
	value := uint64(0)
	reflect.TypeOf(value)
}
