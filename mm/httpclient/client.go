package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	httpURL "net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/mm/httpclient/internal/journal"

	"go.uber.org/zap"
)

const (
	DefaultTTL        = time.Second * 10
	DefaultRetryTimes = 3
	DefaultRetryDelay = time.Millisecond * 100
)

// TODO the retry code may not correct, missing or redundant to be modified in actual use.
func shouldRetry(ctx context.Context, httpCode int) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	switch httpCode {
	case
		_StatusDoReqErr,    // customize
		_StatusReadRespErr, // customize

		http.StatusRequestTimeout,
		http.StatusLocked,
		http.StatusTooEarly,
		http.StatusTooManyRequests,

		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:

		return true

	default:
		return false
	}
}

func Get(url string, form httpURL.Values, options ...Option) (body []byte, header http.Header, statusCode int, err error) {
	return withoutBody(http.MethodGet, url, form, options...)
}

func Delete(url string, form httpURL.Values, options ...Option) (body []byte, header http.Header, statusCode int, err error) {
	return withoutBody(http.MethodDelete, url, form, options...)
}

func PostNoBody(url string, form httpURL.Values, options ...Option) (body []byte, header http.Header, statusCode int, err error) {
	return withoutBody(http.MethodPost, url, form, options...)
}

func PutNoBody(url string, form httpURL.Values, options ...Option) (body []byte, header http.Header, statusCode int, err error) {
	return withoutBody(http.MethodPut, url, form, options...)
}

func PatchNoBody(url string, form httpURL.Values, options ...Option) (body []byte, header http.Header, statusCode int, err error) {
	return withoutBody(http.MethodPatch, url, form, options...)
}

func withoutBody(method, url string, form httpURL.Values, options ...Option) (body []byte, header http.Header, statusCode int, err error) {
	if url = strings.TrimSpace(url); url == "" {
		return nil, nil, -1, errors.New("url required")
	}

	if len(form) > 0 {
		if url, err = AddFormValuesIntoURL(url, form); err != nil {
			return
		}
	}

	ts := time.Now()

	opt := newOption()
	defer func() {
		if opt.Journal != nil {
			opt.Journal.Success = err == nil
			opt.Journal.CostSeconds = time.Since(ts).Seconds()

			if opt.Logger != nil && opt.PrintJournal {
				if err == nil {
					opt.Logger.Info(opt.Desc, zap.Any("journal", marshalJournal(opt.Journal)))
				} else {
					opt.Logger.Error(opt.Desc, zap.Any("journal", marshalJournal(opt.Journal)))
				}
			}

			if opt.NonDurableLogger != nil {
				if err == nil {
					opt.NonDurableLogger.Info(opt.Desc, zap.Any("journal", marshalJournal(opt.Journal)))
				} else {
					opt.NonDurableLogger.Error(opt.Desc, zap.Any("journal", marshalJournal(opt.Journal)))
				}
			}
		}
	}()

	for _, f := range options {
		f(opt)
	}
	opt.Header["Content-Type"] = "application/x-www-form-urlencoded; charset=utf-8"
	if opt.Journal != nil {
		opt.Header[journal.JournalHeader] = opt.Journal.ID
	}

	ttl := opt.TTL
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	background := context.Background()
	if opt.Ctx != nil {
		background = opt.Ctx
	}

	ctx, cancel := context.WithTimeout(background, ttl)
	defer cancel()

	if opt.Journal != nil {
		opt.Journal.Request = &journal.Request{
			TTL:        ttl.String(),
			Method:     method,
			DecodedURL: QueryUnescape(url),
			Header:     opt.Header,
		}
	}

	retryTimes := opt.RetryTimes
	if retryTimes <= 0 {
		retryTimes = DefaultRetryTimes
	}

	retryDelay := opt.RetryDelay
	if retryDelay <= 0 {
		retryDelay = DefaultRetryDelay
	}

	for k := 0; k < retryTimes; k++ {
		body, header, statusCode, err = doHTTP(ctx, method, url, nil, opt)
		if shouldRetry(ctx, statusCode) {
			time.Sleep(retryDelay)
			continue
		}

		return
	}
	return
}

func PostFormBody(url string, form httpURL.Values, options ...Option) (body []byte, header http.Header, statusCode int, err error) {
	return withFormBody(http.MethodPost, url, form, options...)
}

func PostJSONBody(url string, raw json.RawMessage, options ...Option) (body []byte, header http.Header, statusCode int, err error) {
	return withJSONBody(http.MethodPost, url, raw, options...)
}

func PutFormBody(url string, form httpURL.Values, options ...Option) (body []byte, header http.Header, statusCode int, err error) {
	return withFormBody(http.MethodPut, url, form, options...)
}

func PutJSONBody(url string, raw json.RawMessage, options ...Option) (body []byte, header http.Header, statusCode int, err error) {
	return withJSONBody(http.MethodPut, url, raw, options...)
}

func PatchFromBody(url string, form httpURL.Values, options ...Option) (body []byte, header http.Header, statusCode int, err error) {
	return withFormBody(http.MethodPatch, url, form, options...)
}

func PatchJSONBody(url string, raw json.RawMessage, options ...Option) (body []byte, header http.Header, statusCode int, err error) {
	return withJSONBody(http.MethodPatch, url, raw, options...)
}

func withFormBody(method, url string, form httpURL.Values, options ...Option) (body []byte, header http.Header, statusCode int, err error) {
	if url = strings.TrimSpace(url); url == "" {
		return nil, nil, -1, errors.New("url required")
	}
	if len(form) == 0 {
		return nil, nil, -1, errors.New("form required")
	}

	ts := time.Now()

	opt := newOption()
	defer func() {
		if opt.Journal != nil {
			opt.Journal.Success = err == nil
			opt.Journal.CostSeconds = time.Since(ts).Seconds()

			if opt.Logger != nil && opt.PrintJournal {
				if err == nil {
					opt.Logger.Info(opt.Desc, zap.Any("journal", marshalJournal(opt.Journal)))
				} else {
					opt.Logger.Error(opt.Desc, zap.Any("journal", marshalJournal(opt.Journal)))
				}
			}

			if opt.NonDurableLogger != nil {
				if err == nil {
					opt.NonDurableLogger.Info(opt.Desc, zap.Any("journal", marshalJournal(opt.Journal)))
				} else {
					opt.NonDurableLogger.Error(opt.Desc, zap.Any("journal", marshalJournal(opt.Journal)))
				}
			}
		}
	}()

	for _, f := range options {
		f(opt)
	}

	if len(opt.QueryForm) > 0 {
		if url, err = AddFormValuesIntoURL(url, opt.QueryForm); err != nil {
			return
		}
	}

	opt.Header["Content-Type"] = "application/x-www-form-urlencoded; charset=utf-8"
	if opt.Journal != nil {
		opt.Header[journal.JournalHeader] = opt.Journal.ID
	}

	ttl := opt.TTL
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	background := context.Background()
	if opt.Ctx != nil {
		background = opt.Ctx
	}

	ctx, cancel := context.WithTimeout(background, ttl)
	defer cancel()

	formValue := form.Encode()
	if opt.Journal != nil {
		opt.Journal.Request = &journal.Request{
			TTL:        ttl.String(),
			Method:     method,
			DecodedURL: QueryUnescape(url),
			Header:     opt.Header,
			Body:       QueryUnescape(formValue),
		}
	}

	retryTimes := opt.RetryTimes
	if retryTimes <= 0 {
		retryTimes = DefaultRetryTimes
	}

	retryDelay := opt.RetryDelay
	if retryDelay <= 0 {
		retryDelay = DefaultRetryDelay
	}

	for k := 0; k < retryTimes; k++ {
		body, header, statusCode, err = doHTTP(ctx, method, url, []byte(formValue), opt)
		if shouldRetry(ctx, statusCode) {
			time.Sleep(retryDelay)
			continue
		}

		return
	}
	return
}

func withJSONBody(method, url string, raw json.RawMessage, options ...Option) (body []byte, header http.Header, statusCode int, err error) {
	if url = strings.TrimSpace(url); url == "" {
		return nil, nil, -1, errors.New("url required")
	}
	if len(raw) == 0 {
		return nil, nil, -1, errors.New("raw required")
	}

	ts := time.Now()

	opt := newOption()
	defer func() {
		if opt.Journal != nil {
			opt.Journal.Success = err == nil
			opt.Journal.CostSeconds = time.Since(ts).Seconds()

			if opt.Logger != nil && opt.PrintJournal {
				if err == nil {
					opt.Logger.Info(opt.Desc, zap.Any("journal", marshalJournal(opt.Journal)))
				} else {
					opt.Logger.Error(opt.Desc, zap.Any("journal", marshalJournal(opt.Journal)))
				}
			}

			if opt.NonDurableLogger != nil {
				if err == nil {
					opt.NonDurableLogger.Info(opt.Desc, zap.Any("journal", marshalJournal(opt.Journal)))
				} else {
					opt.NonDurableLogger.Error(opt.Desc, zap.Any("journal", marshalJournal(opt.Journal)))
				}
			}
		}
	}()

	for _, f := range options {
		f(opt)
	}

	if len(opt.QueryForm) > 0 {
		if url, err = AddFormValuesIntoURL(url, opt.QueryForm); err != nil {
			return
		}
	}

	opt.Header["Content-Type"] = "application/json; charset=utf-8"
	if opt.Journal != nil {
		opt.Header[journal.JournalHeader] = opt.Journal.ID
	}

	ttl := opt.TTL
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	background := context.Background()
	if opt.Ctx != nil {
		background = opt.Ctx
	}

	ctx, cancel := context.WithTimeout(background, ttl)
	defer cancel()

	if opt.Journal != nil {
		opt.Journal.Request = &journal.Request{
			TTL:        ttl.String(),
			Method:     method,
			DecodedURL: QueryUnescape(url),
			Header:     opt.Header,
			Body:       string(raw),
		}
	}

	retryTimes := opt.RetryTimes
	if retryTimes <= 0 {
		retryTimes = DefaultRetryTimes
	}

	retryDelay := opt.RetryDelay
	if retryDelay <= 0 {
		retryDelay = DefaultRetryDelay
	}

	for k := 0; k < retryTimes; k++ {
		body, header, statusCode, err = doHTTP(ctx, method, url, raw, opt)
		if shouldRetry(ctx, statusCode) {
			time.Sleep(retryDelay)
			continue
		}

		return
	}
	return
}

func PostMultipartFile(url string, payload [][]byte, options ...Option) (body []byte, header http.Header, statusCode int, err error) {
	return withMultipartFile(http.MethodPost, url, payload, options...)
}

func PutMultipartFile(url string, payload [][]byte, options ...Option) (body []byte, header http.Header, statusCode int, err error) {
	return withMultipartFile(http.MethodPut, url, payload, options...)
}

func PatchMultipartFile(url string, payload [][]byte, options ...Option) (body []byte, header http.Header, statusCode int, err error) {
	return withMultipartFile(http.MethodPatch, url, payload, options...)
}

func withMultipartFile(method, url string, payload [][]byte, options ...Option) (body []byte, header http.Header, statusCode int, err error) {
	if url = strings.TrimSpace(url); url == "" {
		return nil, nil, -1, errors.New("url required")
	}
	if len(payload) == 0 {
		return nil, nil, -1, errors.New("payload required")
	}
	for i := range payload {
		if len(payload[i]) == 0 {
			return nil, nil, -1, errors.Errorf("payload[%d] required", i)
		}
	}

	ts := time.Now()

	opt := newOption()
	defer func() {
		if opt.Journal != nil {
			opt.Journal.Success = err == nil
			opt.Journal.CostSeconds = time.Since(ts).Seconds()

			if opt.Logger != nil && opt.PrintJournal {
				if err == nil {
					opt.Logger.Info(opt.Desc, zap.Any("journal", marshalJournal(opt.Journal)))
				} else {
					opt.Logger.Error(opt.Desc, zap.Any("journal", marshalJournal(opt.Journal)))
				}
			}

			if opt.NonDurableLogger != nil {
				if err == nil {
					opt.NonDurableLogger.Info(opt.Desc, zap.Any("journal", marshalJournal(opt.Journal)))
				} else {
					opt.NonDurableLogger.Error(opt.Desc, zap.Any("journal", marshalJournal(opt.Journal)))
				}
			}
		}
	}()

	for _, f := range options {
		f(opt)
	}

	if len(opt.QueryForm) > 0 {
		if url, err = AddFormValuesIntoURL(url, opt.QueryForm); err != nil {
			return
		}
	}

	buf := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(buf)

	for i := range payload {
		file, err := writer.CreateFormFile(strconv.Itoa(i), strconv.Itoa(i))
		if err != nil {
			return nil, nil, -1, errors.Wrap(err, "create multipart file err")
		}

		if _, err := file.Write(payload[i]); err != nil {
			return nil, nil, -1, errors.Wrap(err, "write multipart file err")
		}
	}

	if err := writer.Close(); err != nil {
		return nil, nil, -1, errors.Wrap(err, "close multipart file err")
	}

	opt.Header["Content-Type"] = writer.FormDataContentType()
	if opt.Journal != nil {
		opt.Header[journal.JournalHeader] = opt.Journal.ID
	}

	ttl := opt.TTL
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	background := context.Background()
	if opt.Ctx != nil {
		background = opt.Ctx
	}

	ctx, cancel := context.WithTimeout(background, ttl)
	defer cancel()

	if opt.Journal != nil {
		opt.Journal.Request = &journal.Request{
			TTL:        ttl.String(),
			Method:     method,
			DecodedURL: QueryUnescape(url),
			Header:     opt.Header,
		}
	}

	retryTimes := opt.RetryTimes
	if retryTimes <= 0 {
		retryTimes = DefaultRetryTimes
	}

	retryDelay := opt.RetryDelay
	if retryDelay <= 0 {
		retryDelay = DefaultRetryDelay
	}

	for k := 0; k < retryTimes; k++ {
		body, header, statusCode, err = doHTTP(ctx, method, url, buf.Bytes(), opt)
		if shouldRetry(ctx, statusCode) {
			time.Sleep(retryDelay)
			continue
		}

		return
	}
	return
}
