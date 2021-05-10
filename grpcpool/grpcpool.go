package grpcpool

import (
	"context"
	stderr "errors"
	"sync"
	"time"

	"github.com/byepichi/pkg/crypto"
	"github.com/byepichi/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

const (
	// defaultMaxStreamsPerStub used grpc/internal/transport.defaultMaxStreamsClient
	// use grpc.MaxConcurrentStreams() to change the default config
	defaultMaxStreamsPerStub = 100
	defaultStubIdleSeconds   = float64(time.Minute / time.Second)
	defaultRecycleTicker     = time.Second * 5
	defaultChannelCapacity   = 10

	namespace = "byepichi"
	subsystem = "grpcpool"
)

var (
	// ClosedErr pool has closed
	ClosedErr = stderr.New("grpc pool has closed")

	counters = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "counters",
		Help:      "get/restore stub(s) from/into pool",
	}, []string{"method"})

	connections = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections",
		Help:      "grpc connections",
	}, []string{"addr"})

	storages = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "storages",
		Help:      "stub is full, wait to restore",
	}, []string{"addr"})

	cost = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "cost",
		Help:      "the cost of get/restore stub",
	}, []string{"method"})
)

// Builder builer for create grpc client connection
type Builder func() (*grpc.ClientConn, error)

var _ ClientConn = (*clientConn)(nil)

// ClientConn wrapper for *grpc.ClientConn
type ClientConn interface {
	grpc.ClientConnInterface
	t()
}

type clientConn struct {
	*grpc.ClientConn
}

func (*clientConn) t() {}

var _ Stub = (*stub)(nil)

// Stub the real handler in pool
type Stub interface {
	Conn() ClientConn
	t()
}

type stub struct {
	id      string
	conn    *clientConn
	streams int32
	ts      time.Time
}

func newStubP(builder Builder) *stub {
	conn, err := builder()
	if err != nil {
		panic(err)
	}

	return &stub{
		id:   crypto.A20RID(crypto.IDPrefix{'S', 'T'}),
		conn: &clientConn{conn},
	}
}

func (s *stub) Conn() ClientConn {
	return s.conn
}

func (s *stub) t() {}

// Option optional configs
type Option func(*option)

type option struct {
	enablePrometheus bool
}

// WithEnablePrometheus enable to record prometheus metrics
func WithEnablePrometheus() Option {
	return func(opt *option) {
		opt.enablePrometheus = true
	}
}

var _ Pool = (*pool)(nil)

// Pool grpc pool
type Pool interface {
	Close()
	Get() (Stub, error)
	Restore(Stub)
}

type pool struct {
	sync.WaitGroup
	ctx              context.Context
	cancel           context.CancelFunc
	ticker           *time.Ticker
	builder          Builder
	pool             *stack
	storage          map[string]*stub
	toRecycle        map[string]time.Time
	enablePrometheus bool
	restore          chan *stub
	buffer           chan *stub
	activity         chan byte
	closed           chan bool
}

// NewPool create a grpc pool
func NewPool(builder Builder, options ...Option) (Pool, error) {
	if builder == nil {
		return nil, errors.New("builder required")
	}

	opt := new(option)
	for _, f := range options {
		f(opt)
	}

	if opt.enablePrometheus {
		prometheus.MustRegister(counters)
		prometheus.MustRegister(connections)
		prometheus.MustRegister(storages)
		prometheus.MustRegister(cost)
	}

	stack := new(stack)
	stack.Push(newStubP(builder))

	ctx, cancel := context.WithCancel(context.Background())
	pool := &pool{
		ctx:              ctx,
		cancel:           cancel,
		ticker:           time.NewTicker(defaultRecycleTicker),
		builder:          builder,
		pool:             stack,
		storage:          make(map[string]*stub),
		toRecycle:        make(map[string]time.Time),
		enablePrometheus: opt.enablePrometheus,
		restore:          make(chan *stub, defaultChannelCapacity),
		buffer:           make(chan *stub, defaultChannelCapacity),
		activity:         make(chan byte, defaultChannelCapacity),
		closed:           make(chan bool),
	}

	go pool.handler()
	return pool, nil
}

func (p *pool) Close() {
	select {
	case <-p.ctx.Done():
	default:
		p.cancel()
		p.ticker.Stop()

		p.Wait()
		close(p.closed)

		for id := range p.toRecycle {
			p.pool.Remove(id).conn.Close()
		}

		for !p.pool.Empty() {
			p.pool.Pop().conn.Close()
		}
	}
}

func (p *pool) handler() {
	for {
		select {
		case <-p.closed:
			return

		case <-p.ticker.C:
			for id, ts := range p.toRecycle {
				if time.Since(ts).Seconds() > defaultStubIdleSeconds {
					if len(p.toRecycle) != 1 {
						stub := p.pool.Remove(id)
						stub.conn.Close()
						if p.enablePrometheus {
							connections.WithLabelValues(stub.conn.Target()).Sub(1)
						}
					}
					delete(p.toRecycle, id)
				}
			}

		case stub := <-p.restore:
			if s, ok := p.storage[stub.id]; ok {
				delete(p.storage, s.id)
				s.streams--
				p.pool.Push(s)

				if p.enablePrometheus {
					counters.WithLabelValues("restores").Add(1)
					storages.WithLabelValues(s.conn.Target()).Sub(1)
				}
				continue
			}

			if stub.streams >= 1 {
				if stub.streams--; stub.streams == 0 {
					p.toRecycle[stub.id] = time.Now()
				}

				if p.enablePrometheus {
					counters.WithLabelValues("restores").Add(1)
				}
			}

		case <-p.activity:
		GET:
			stub := p.pool.Peek()
			if stub.streams+1 <= defaultMaxStreamsPerStub {
				if stub.streams++; stub.streams == 1 {
					delete(p.toRecycle, stub.id)
				}

				if p.enablePrometheus {
					counters.WithLabelValues("gets").Add(1)
				}

				goto PUT
			}

			p.storage[stub.id] = p.pool.Pop()
			if p.enablePrometheus {
				storages.WithLabelValues(stub.conn.Target()).Add(1)
			}
			if !p.pool.Empty() {
				goto GET
			}

			stub = newStubP(p.builder)
			stub.streams++
			p.pool.Push(stub)

			if p.enablePrometheus {
				counters.WithLabelValues("gets").Add(1)
				connections.WithLabelValues(stub.conn.Target()).Add(1)
			}

		PUT:
			p.buffer <- stub
		}
	}
}

func (p *pool) Get() (Stub, error) {
	ts := time.Now()

	for {
		select {
		case <-p.ctx.Done():
			return nil, ClosedErr

		default:
			p.activity <- 1
			if stub := <-p.buffer; stub.conn.GetState() != connectivity.Shutdown {
				p.Add(1)
				if p.enablePrometheus {
					cost.WithLabelValues("get").Set(time.Since(ts).Seconds())
				}
				return stub, nil
			}
		}
	}
}

func (p *pool) Restore(s Stub) {
	if s == nil {
		return
	}

	ts := time.Now()
	p.restore <- s.(*stub)
	if p.enablePrometheus {
		cost.WithLabelValues("restore").Set(time.Since(ts).Seconds())
	}
	p.Done()
}
