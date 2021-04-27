package rate

import (
	"crypto/tls"
	"time"

	"github.com/byepichi/pkg/errors"

	"github.com/go-redis/redis/v7"
)

// NewSingleRedis create a single redis client
func NewSingleRedis(endpoint, password string, db uint16, tls *tls.Config) (Redis, error) {
	redis := redis.NewClient(&redis.Options{
		Addr:               endpoint,
		Password:           password,
		DB:                 int(db),
		MaxRetries:         3,
		DialTimeout:        time.Second * 2,
		ReadTimeout:        time.Millisecond * 200,
		WriteTimeout:       time.Millisecond * 200,
		PoolSize:           5,
		MinIdleConns:       3,
		MaxConnAge:         time.Minute,
		PoolTimeout:        time.Minute,
		IdleTimeout:        time.Second * 30,
		IdleCheckFrequency: time.Second * 2,
		TLSConfig:          tls,
	})

	if _, err := redis.Ping().Result(); err != nil {
		return nil, errors.Wrap(err, "redis ping err")
	}

	return redis, nil
}

// NewClusterRedis create a cluster redis client
func NewClusterRedis(endpoints []string, password string, tls *tls.Config) (Redis, error) {
	redis := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:              endpoints,
		Password:           password,
		MaxRetries:         3,
		DialTimeout:        time.Second * 2,
		ReadTimeout:        time.Millisecond * 200,
		WriteTimeout:       time.Millisecond * 200,
		PoolSize:           5,
		MinIdleConns:       3,
		MaxConnAge:         time.Minute,
		PoolTimeout:        time.Minute,
		IdleTimeout:        time.Second * 30,
		IdleCheckFrequency: time.Second * 2,
		TLSConfig:          tls,
	})

	if _, err := redis.Ping().Result(); err != nil {
		return nil, errors.Wrap(err, "redis ping err")
	}

	return redis, nil
}
