package channel

import (
	"context"
	stderr "errors"
	"reflect"
	"sync"

	"github.com/bluekaki/pkg/errors"
)

var _ Channel = (*channel)(nil)

// ErrClosed returned if Append called after Close()
var ErrClosed = stderr.New("channel has closed, do not recevie value anymore.")

// Channel a graceful-close channel
type Channel interface {
	// Close will block until all data in channel consumed completely
	Close()
	// T return channel's typeof
	T() string
	// Append value into channel, the value must be format of typeof defined in NewChannel
	Append(value interface{}) error
}

// Consumer define a handle how consume data from channel
type Consumer = func(value interface{})

type config struct {
	workers int
}

// Option how setup consumer
type Option func(*config)

// WithWorker setup the number of consumers
func WithWorker(num uint16) Option {
	return func(conf *config) {
		conf.workers = int(num)
	}
}

type channel struct {
	ctx      context.Context
	cancel   context.CancelFunc
	bufferWG *sync.WaitGroup
	workerWG *sync.WaitGroup
	typeOf   reflect.Type
	buffer   chan interface{}
	consumer Consumer
}

// NewChannel create a new graceful-close channel instance, data in channel must be typeof;
// with a single consumer by default.
func NewChannel(typeOf reflect.Type, capactiy uint16, consumer Consumer, opts ...Option) (Channel, error) {
	if typeOf == nil {
		return nil, errors.New("typeOf required")
	}
	if consumer == nil {
		return nil, errors.New("consumer required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	channel := &channel{
		ctx:      ctx,
		cancel:   cancel,
		bufferWG: new(sync.WaitGroup),
		workerWG: new(sync.WaitGroup),
		typeOf:   typeOf,
		buffer:   make(chan interface{}, int(capactiy)),
		consumer: consumer,
	}

	conf := new(config)
	for _, opt := range opts {
		opt(conf)
	}

	if conf.workers == 0 {
		conf.workers = 1
	}
	channel.consume(conf)

	return channel, nil
}

func (c *channel) T() string {
	return c.typeOf.String()
}

func (c *channel) Close() {
	select {
	case <-c.ctx.Done():
	default:
		c.cancel()
		c.bufferWG.Wait()
		close(c.buffer)

		c.workerWG.Wait()
	}
}

func (c *channel) Append(value interface{}) error {
	if value == nil {
		return errors.New("value required")
	}
	if c.typeOf != reflect.TypeOf(value) {
		return errors.Errorf("value must be type of %s", c.typeOf.String())
	}

	select {
	case <-c.ctx.Done():
		return ErrClosed
	default:
	}

	c.bufferWG.Add(1)
	c.buffer <- value
	c.bufferWG.Done()

	return nil
}

func (c *channel) consume(conf *config) {
	for k := 0; k < conf.workers; k++ {
		c.workerWG.Add(1)
		go func() {
			defer c.workerWG.Done()

			for value := range c.buffer {
				c.consumer(value)
			}
		}()
	}
}
