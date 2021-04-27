package rate

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/byepichi/pkg/errors"

	"github.com/go-redis/redis/v7"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

var _ Limiter = (*limiter)(nil)

const (
	prefix                = "byepichi-ratelimiter:"
	defaultTickerInterval = time.Millisecond * 300
)

// Config how setup Limiter
type Config struct {
	// Identifier a distinguish between different limiters
	Identifier string
	// TickerInterval duration of report/fetch increment to/from redis
	TickerInterval time.Duration
	// Limit the expected rate
	Limit struct {
		// Upper the expected maximum speed
		Upper uint32
		// Lower used when lose redis
		Lower uint16
	}
	// Log setup logger, if enable and no logger set, zap.NewProduction() will used.
	Log struct {
		Enable bool
		Logger *zap.Logger
	}
}

// Redis a single or cluster redis instance
type Redis interface {
	// Close the redis
	Close() error
	// IncrBy incr by value
	IncrBy(key string, value int64) *redis.IntCmd
	// Get by key
	Get(key string) *redis.StringCmd
}

// Limiter a distributed speed limiter
type Limiter interface {
	// Close the limiter
	Close() error
	// UpdateRate dynamic update rate, err will returned if parameters set zero
	UpdateRate(limitUpper uint32, limitLower uint16) error
	// Allow whether event can happen at time now
	Allow()
}

type limiter struct {
	ctx        context.Context
	cancel     context.CancelFunc
	logger     *zap.Logger
	mux        *sync.Mutex
	redis      Redis
	identifier string
	limitUpper int
	limitLower int
	limiter    *rate.Limiter
	summary    uint64
}

// NewLimiter create a new instance of Limiter
func NewLimiter(conf *Config, redis Redis) (Limiter, error) {
	if conf == nil {
		return nil, errors.New("conf required")
	}
	if conf.Identifier == "" {
		return nil, errors.New("conf.Identifier required")
	}
	if conf.Limit.Upper == 0 {
		return nil, errors.New("conf.Limit.Upper required")
	}
	if conf.Limit.Lower == 0 {
		return nil, errors.New("conf.Limit.Lower required")
	}
	if redis == nil {
		return nil, errors.New("redis required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	limiter := &limiter{
		ctx:        ctx,
		cancel:     cancel,
		logger:     conf.Log.Logger,
		mux:        new(sync.Mutex),
		redis:      redis,
		identifier: prefix + conf.Identifier,
		limitUpper: int(conf.Limit.Upper),
		limitLower: int(conf.Limit.Lower),
		limiter:    rate.NewLimiter(rate.Limit(1), 1), // slow start
	}

	if conf.Log.Logger == nil && conf.Log.Enable {
		// just disable log if err occurs
		if logger, err := zap.NewProduction(); err == nil {
			limiter.logger = logger
		}
	}

	go func() {
		time.Sleep(time.Second * 5) // avoid rate increase rapidly during multi instances startup
		limiter.accelerate()
	}()

	tickerInterval := conf.TickerInterval
	if tickerInterval == 0 {
		tickerInterval = defaultTickerInterval
	}

	go limiter.report(tickerInterval)
	go limiter.fetch(tickerInterval)

	return limiter, nil
}

func (l *limiter) accelerate() {
	l.mux.Lock()
	defer l.mux.Unlock()

	l.limiter.SetLimit(rate.Limit(l.limitUpper))
}

func (l *limiter) decelerate() {
	l.mux.Lock()
	defer l.mux.Unlock()

	l.limiter.SetLimit(rate.Limit(l.limitLower))
}

func (l *limiter) Close() error {
	select {
	case <-l.ctx.Done():
	default:
		l.cancel()
		if err := l.redis.Close(); err != nil {
			return errors.Wrap(err, "close redis err")
		}
	}
	return nil
}

func (l *limiter) UpdateRate(limitUpper uint32, limitLower uint16) error {
	if limitUpper == 0 {
		return errors.New("limitUpper required")
	}
	if limitLower == 0 {
		return errors.New("limitLower required")
	}

	l.mux.Lock()
	l.limitUpper = int(limitUpper)
	l.limitLower = int(limitLower)
	l.mux.Unlock()

	l.accelerate()
	return nil
}

func (l *limiter) Allow() {
	if !l.limiter.Allow() {
		time.Sleep(l.limiter.Reserve().Delay())
	}
	atomic.AddUint64(&l.summary, 1)
}

// report increment to redis
func (l *limiter) report(tickerInterval time.Duration) {
	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	var last uint64
	do := func() {
		now := atomic.LoadUint64(&l.summary)
		if now == last {
			return
		}

		if err := l.redis.IncrBy(l.identifier, int64(now-last)).Err(); err != nil && l.logger != nil {
			l.logger.Error("redis.IncrBy err", zap.String("key", l.identifier), zap.Error(err))
		}
		last = now
	}

	for {
		select {
		case <-l.ctx.Done():
			return
		case <-ticker.C:
			do()
		}
	}
}

// fetch increment from redis
func (l *limiter) fetch(tickerInterval time.Duration) {
	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	const (
		initMode = iota
		safeMode
		normalMode
	)
	mode := initMode

	var last uint64
	do := func() {
		now, err := l.redis.Get(l.identifier).Uint64()
		if err != nil { // first time may err cause key not exist
			if l.logger != nil {
				l.logger.Error("redis.Get err", zap.String("key", l.identifier), zap.Error(err))
			}

			if mode != safeMode {
				l.decelerate()
				mode = safeMode
			}
			return
		}

		if now != last {
			if mode != normalMode {
				l.accelerate()
				mode = normalMode
			}
			l.limiter.ReserveN(time.Now(), int(now-last))
			last = now
		}
	}

	for {
		select {
		case <-l.ctx.Done():
			return
		case <-ticker.C:
			do()
		}
	}
}
