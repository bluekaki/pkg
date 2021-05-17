package httpclient

import (
	"context"
	"time"

	"github.com/bluekaki/pkg/mm/internal/journal"

	"go.uber.org/zap"
)

// Journal 记录内部流转信息
type Journal = journal.T

// Option 自定义设置http请求
type Option func(*option)

type option struct {
	Ctx            context.Context
	TTL            time.Duration
	Header         map[string]string
	Journal        *journal.Journal
	Dialog         *journal.Dialog
	Logger         *zap.Logger
	RetryTimes     int
	RetryDelay     time.Duration
	PrintJournal   bool
	MarshalJournal bool
	Desc           string
}

func newOption() *option {
	return &option{
		Header: make(map[string]string),
	}
}

// WithContext ttl会基于此ctx做计时
func WithContext(ctx context.Context) Option {
	return func(opt *option) {
		opt.Ctx = ctx
	}
}

// WithTTL 本次http请求最长执行时间
func WithTTL(ttl time.Duration) Option {
	return func(opt *option) {
		opt.TTL = ttl
	}
}

// WithHeader 设置http header，可以调用多次设置多对key-value
func WithHeader(key, value string) Option {
	return func(opt *option) {
		opt.Header[key] = value
	}
}

// WithNewJournal 创建新的Journal用于记录
func WithNewJournal(id string) Option {
	return WithJournal(journal.NewJournal(id))
}

// WithJournal 设置Journal以便记录内部流转信息
func WithJournal(j Journal) Option {
	return func(opt *option) {
		if j != nil {
			opt.Journal = j.(*journal.Journal)
			opt.Dialog = new(journal.Dialog)
		}
	}
}

// WithLogger 设置logger以便打印关键日志
func WithLogger(logger *zap.Logger) Option {
	return func(opt *option) {
		opt.Logger = logger
	}
}

// WithRetryTimes 如果请求失败，最多重试N次
func WithRetryTimes(retryTimes int) Option {
	return func(opt *option) {
		opt.RetryTimes = retryTimes
	}
}

// WithRetryDelay 在重试前，延迟等待一会
func WithRetryDelay(retryDelay time.Duration) Option {
	return func(opt *option) {
		opt.RetryDelay = retryDelay
	}
}

// WithPrintJournal 打印journal
func WithPrintJournal(desc string) Option {
	return func(opt *option) {
		opt.PrintJournal = true
		opt.Desc = desc
	}
}

// WithMarshalJournal 序列化journal
func WithMarshalJournal() Option {
	return func(opt *option) {
		opt.MarshalJournal = true
	}
}
