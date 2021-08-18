package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/mm/internal/journal"

	"go.uber.org/zap"
)

const (
	// _StatusReadRespErr read resp body err, should re-call doHTTP again.
	_StatusReadRespErr = -204
	// _StatusDoReqErr do req err, should re-call doHTTP again.
	_StatusDoReqErr = -500
)

var defaultClient = &http.Client{
	Transport: &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func doHTTP(ctx context.Context, method, url string, payload []byte, opt *option) ([]byte, http.Header, int, error) {
	ts := time.Now()

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(payload))
	if err != nil {
		return nil, nil, -1, errors.Wrapf(err, "new request [%s %s] err", method, url)
	}

	for key, value := range opt.Header {
		req.Header.Set(key, value)
	}

	resp, err := defaultClient.Do(req)
	if err != nil {
		err = errors.Wrapf(err, "do request [%s %s] err", method, url)
		if opt.Dialog != nil {
			opt.Dialog.AppendResponse(&journal.Response{
				Body:        err.Error(),
				CostSeconds: time.Since(ts).Seconds(),
			})
		}

		if opt.Logger != nil {
			opt.Logger.Warn("doHTTP got err", zap.Error(err))
		}
		return nil, nil, _StatusDoReqErr, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.Wrapf(err, "read resp body from [%s %s] err", method, url)
		if opt.Dialog != nil {
			opt.Dialog.AppendResponse(&journal.Response{
				Body:        err.Error(),
				CostSeconds: time.Since(ts).Seconds(),
			})
		}

		if opt.Logger != nil {
			opt.Logger.Warn("doHTTP got err", zap.Error(err))
		}
		return nil, nil, _StatusReadRespErr, err
	}

	defer func() {
		if opt.Dialog != nil {
			var raw string
			if strings.Contains(resp.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
				raw = QueryUnescape(string(body)) // unsafe
			} else {
				raw = string(body) // unsafe
			}

			opt.Dialog.AppendResponse(&journal.Response{
				Header:      journal.ToJournalHeader(resp.Header),
				StatusCode:  resp.StatusCode,
				Status:      resp.Status,
				Body:        raw,
				CostSeconds: time.Since(ts).Seconds(),
			})
		}
	}()

	if resp.StatusCode != http.StatusOK {
		err = newReplyErr(
			resp.StatusCode,
			errors.Errorf("do [%s %s] return code: %d message: %s", method, url, resp.StatusCode, string(body)),
		)
	}

	return body, resp.Header, resp.StatusCode, err
}

// addFormValuesIntoURL append url.Values into url string
func addFormValuesIntoURL(rawURL string, form url.Values) (string, error) {
	if rawURL == "" {
		return "", errors.New("rawURL required")
	}
	if len(form) == 0 {
		return "", errors.New("form required")
	}

	target, err := url.Parse(rawURL)
	if err != nil {
		return "", errors.Wrapf(err, "parse rawURL `%s` err", rawURL)
	}

	urlValues := target.Query()
	for key, values := range form {
		for _, value := range values {
			urlValues.Add(key, value)
		}
	}

	target.RawQuery = urlValues.Encode()
	return target.String(), nil
}

func QueryUnescape(uri string) string {
	decodedUri, err := url.QueryUnescape(uri)
	if err != nil {
		return uri
	}

	return decodedUri
}

func marshalJournal(journal *journal.Journal) interface{} {
	raw, _ := json.Marshal(journal)

	if os.Getenv("MarshalJournal") == "true" {
		return string(raw)
	}
	return json.RawMessage(raw)
}
