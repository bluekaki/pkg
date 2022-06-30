package pushgateway

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bluekaki/pkg/errors"

	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

type Option func(*option)

type option struct {
	labels map[string]string
	ttl    time.Duration
}

func WithLabel(name, value string) Option {
	return func(opt *option) {
		opt.labels[name] = value
	}
}

func WithTTL(duration time.Duration) Option {
	return func(opt *option) {
		if duration > 0 {
			opt.ttl = duration
		}
	}
}

func Run(localMetrics, remotePushgateway string, opts ...Option) chan error {
	opt := &option{labels: make(map[string]string)}
	for _, f := range opts {
		f(opt)
	}

	if opt.ttl == 0 {
		opt.ttl = time.Second * 10
	}

	lables := make([]string, 0, len(opt.labels))
	for name, value := range opt.labels {
		lables = append(lables, fmt.Sprintf("%s/%s", name, value))
	}

	remotePushgateway = fmt.Sprintf("%s/metrics/job/pushgateway-converter/%s", remotePushgateway, strings.Join(lables, "/"))

	ch := make(chan error, 10)
	go func() {
		notify := func(err error) {
			timer := time.NewTimer(time.Second * 2)
			defer timer.Stop()

			select {
			case <-timer.C:
			case ch <- err:
			}
		}

		ticker := time.NewTicker(time.Millisecond * 10)
		defer ticker.Stop()

		firstTime := true
		for {
			<-ticker.C
			if firstTime {
				firstTime = false
				ticker.Reset(time.Second * 10)
			}

			if err := fetchAndPush(localMetrics, remotePushgateway, opt.ttl); err != nil {
				notify(err)
			}
		}
	}()

	return ch
}

var defaultClient = &http.Client{
	Transport: &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func fetchAndPush(localMetrics, remotePushgateway string, ttl time.Duration) error {
	resp, err := http.Get(localMetrics)
	if err != nil {
		return errors.Wrapf(err, "get metrics from %s err", localMetrics)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return errors.Errorf("get metrics from %s err, status: %s, message: %s", localMetrics, resp.Status, string(body))
	}

	decoder := expfmt.NewDecoder(resp.Body, expfmt.FmtProtoText)
	payload := bytes.NewBuffer(nil)

	for {
		mf := new(dto.MetricFamily)
		err := decoder.Decode(mf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "decode textmetrics to metricfamily err")
		}

		if _, err = pbutil.WriteDelimited(payload, mf); err != nil {
			return errors.Wrap(err, "marshal metricfamily err")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, remotePushgateway, payload)
	if err != nil {
		return errors.Wrap(err, "create request to pushgateway err")
	}

	req.Header.Set("content-type", "application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily; encoding=delimited")

	ack, err := defaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "push metrics to pushgateway err")
	}
	defer ack.Body.Close()

	if ack.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(ack.Body)
		return errors.Errorf("push metrics to pushgateway err, status: %s, message: %s", ack.Status, string(body))
	}

	return nil
}
