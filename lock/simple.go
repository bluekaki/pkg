package lock

import (
	"strings"
	"time"

	"github.com/byepichi/pkg/errors"

	"github.com/go-redis/redis/v7"
)

const (
	DefaultRetryTimes = uint8(20)
	DefaultRetryDelay = time.Millisecond * 300
	DefaultLockTTL    = time.Second * 2
)

type RedisClient interface {
	TTL(key string) *redis.DurationCmd
	Del(keys ...string) *redis.IntCmd
	SetNX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd
}

type Handler func() error

type Option func(*option)

type option struct {
	RedisClient RedisClient
	RetryTimes  uint8
	RetryDelay  time.Duration
	LockTTL     time.Duration
}

func WithRedisClient(client RedisClient) Option {
	return func(opt *option) {
		opt.RedisClient = client
	}
}

func WithRetryTimes(times uint8) Option {
	return func(opt *option) {
		if times > 0 {
			opt.RetryTimes = times
		}
	}
}

func WithRetryDelay(delay time.Duration) Option {
	return func(opt *option) {
		if delay > 0 {
			opt.RetryDelay = delay
		}
	}
}

func WithLockTTL(ttl time.Duration) Option {
	return func(opt *option) {
		if ttl > 0 {
			opt.LockTTL = ttl
		}
	}
}

func Simple(key string, handler Handler, options ...Option) error {
	if key = strings.TrimSpace(key); key == "" {
		return errors.New("key required")
	}

	if handler == nil {
		return errors.New("handler required")
	}

	opt := new(option)
	for _, f := range options {
		f(opt)
	}

	if opt.RedisClient == nil {
		return errors.New("redis client required")
	}

	retryTimes := DefaultRetryTimes
	if opt.RetryTimes > 0 {
		retryTimes = opt.RetryTimes
	}

	retryDelay := DefaultRetryDelay
	if opt.RetryDelay > 0 {
		retryDelay = opt.RetryDelay
	}

	lockTTL := DefaultLockTTL
	if opt.LockTTL > 0 {
		lockTTL = opt.LockTTL
	}

	for k := uint8(0); k < retryTimes; k++ {
		ok, err := opt.RedisClient.SetNX(key, "", lockTTL).Result()
		if err != nil {
			return errors.Wrapf(err, "lock setnx key: %s err", key)
		}

		if !ok {
			time.Sleep(retryDelay)
			continue
		}

		defer opt.RedisClient.Del(key)
		if err = handler(); err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	return errors.Errorf("lock failed after %d attempts", retryTimes)
}
