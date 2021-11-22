package httpclient

import (
	"context"
	"net/url"
	"time"

	"github.com/bluekaki/pkg/mm/httpclient/internal/journal"

	"go.uber.org/zap"
)

type Option func(*option)

type option struct {
	Ctx          context.Context
	TTL          time.Duration
	Header       map[string]string
	Journal      *journal.Journal
	RetryTimes   int
	RetryDelay   time.Duration
	PrintJournal bool
	Logger       *zap.Logger
	Desc         string
	QueryForm    url.Values
}

func newOption() *option {
	return &option{
		Header:  make(map[string]string),
		Journal: journal.NewJournal(""),
	}
}

func WithContext(ctx context.Context) Option {
	return func(opt *option) {
		opt.Ctx = ctx
	}
}

func WithTTL(ttl time.Duration) Option {
	return func(opt *option) {
		opt.TTL = ttl
	}
}

func WithHeader(key, value string) Option {
	return func(opt *option) {
		opt.Header[key] = value
	}
}

func WithRetryTimes(retryTimes int) Option {
	return func(opt *option) {
		opt.RetryTimes = retryTimes
	}
}

func WithRetryDelay(retryDelay time.Duration) Option {
	return func(opt *option) {
		opt.RetryDelay = retryDelay
	}
}

// WithJournalID set an unique journal id
func WithJournalID(id string) Option {
	return func(opt *option) {
		opt.Journal.ID = id
	}
}

func WithPrintJournal(logger *zap.Logger, desc string) Option {
	return func(opt *option) {
		opt.PrintJournal = true
		opt.Logger = logger
		opt.Desc = desc
	}
}

// WithQueryForm add some queryform values(only works for with body)
func WithQueryForm(form url.Values) Option {
	return func(opt *option) {
		opt.QueryForm = form
	}
}
