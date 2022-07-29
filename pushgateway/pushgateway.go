package pushgateway

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
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

type Porter interface {
	io.Closer
	Errors() <-chan error
}

type porter struct {
	sync.WaitGroup
	ctx               context.Context
	cancel            context.CancelFunc
	localMetrics      string
	remotePushgateway string
	ttl               time.Duration
	errors            chan error
}

func NewPorter(localMetrics, remotePushgateway string, opts ...Option) Porter {
	if localMetrics = strings.TrimSpace(localMetrics); localMetrics == "" {
		panic("localMetrics required")
	}
	if remotePushgateway = strings.TrimSpace(remotePushgateway); remotePushgateway == "" {
		panic("remotePushgateway required")
	}

	opt := &option{labels: make(map[string]string)}
	for _, f := range opts {
		f(opt)
	}

	if opt.ttl == 0 {
		opt.ttl = time.Second * 10
	}

	opt.labels["_created_at"] = time.Now().Format(time.RFC3339)

	lables := make([]string, 0, len(opt.labels))
	for name, value := range opt.labels {
		lables = append(lables, fmt.Sprintf("%s/%s", name, value))
	}

	sort.Strings(lables)
	remotePushgateway = fmt.Sprintf("%s/metrics/job/metrics-porter/%s", remotePushgateway, strings.Join(lables, "/"))

	ctx, cancel := context.WithCancel(context.Background())
	porter := &porter{
		ctx:               ctx,
		cancel:            cancel,
		localMetrics:      localMetrics,
		remotePushgateway: remotePushgateway,
		ttl:               opt.ttl,
		errors:            make(chan error, 10),
	}

	go porter.run()
	return porter
}

func (p *porter) run() {
	p.Add(1)
	defer p.Done()

	notify := func(err error) {
		ctx, cancel := context.WithTimeout(p.ctx, time.Second*2)
		defer cancel()

		select {
		case <-ctx.Done():
		case p.errors <- err:
		}
	}

	ticker := time.NewTicker(time.Millisecond * 10)
	defer ticker.Stop()

	firstTime := true
	for {
		select {
		case <-p.ctx.Done():
			return

		case <-ticker.C:
			if firstTime {
				firstTime = false
				ticker.Reset(time.Second * 10)
			}

			if err := p.fetchAndPush(); err != nil {
				notify(err)
			}
		}
	}
}

var defaultClient = &http.Client{
	Transport: &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func (p *porter) fetchAndPush() error {
	resp, err := http.Get(p.localMetrics)
	if err != nil {
		return errors.Wrapf(err, "get metrics from %s err", p.localMetrics)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return errors.Errorf("get metrics from %s err, status: %s, message: %s", p.localMetrics, resp.Status, string(body))
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

	ctx, cancel := context.WithTimeout(p.ctx, p.ttl)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.remotePushgateway, payload)
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

func (p *porter) Close() error {
	p.cancel()
	p.Wait()

	ctx, cancel := context.WithTimeout(context.Background(), p.ttl)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, p.remotePushgateway, nil)
	if err != nil {
		return errors.Wrap(err, "create delete request to pushgateway err")
	}

	ack, err := defaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "delete metrics from pushgateway err")
	}
	defer ack.Body.Close()

	if ack.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(ack.Body)
		return errors.Errorf("delete metrics from pushgateway err, status: %s, message: %s", ack.Status, string(body))
	}

	return nil
}

func (p *porter) Errors() <-chan error {
	return p.errors
}
