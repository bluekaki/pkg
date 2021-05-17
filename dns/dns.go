package dns

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/bluekaki/pkg/errors"
)

var _ GoogleDNS = (*googleDNS)(nil)

// GoogleDNS  google public dns
type GoogleDNS interface {
	Close()
	Query(host string) (network string, ip string, err error)
}

type googleDNS struct {
	dummy  bool
	ctx    context.Context
	cancel context.CancelFunc
	cache  *sync.Map
}

type dnsRecord struct {
	host    string
	network string // tcp4 or tcp6
	ip      string // ipv6 in format "[0:0::0]"
	ts      time.Time
}

// NewGoogleDNS create google dns instance
func NewGoogleDNS(dummy ...bool) GoogleDNS {
	ctx, cancel := context.WithCancel(context.Background())

	dns := &googleDNS{
		ctx:    ctx,
		cancel: cancel,
		cache:  new(sync.Map),
	}
	if dummy != nil {
		dns.dummy = true
	}

	go dns.cleaner()
	return dns
}

func (g *googleDNS) Close() {
	g.cancel()
}

func (g *googleDNS) store(host, network, ip string) {
	g.cache.Store(host, &dnsRecord{
		host:    host,
		network: network,
		ip:      ip,
		ts:      time.Now(),
	})
}

func (g *googleDNS) localQuery(host string) (found bool, network, ip string) {
	if value, ok := g.cache.Load(host); ok {
		record := value.(*dnsRecord)

		found = true
		network = record.network
		ip = record.ip
	}
	return
}

func (g *googleDNS) cleaner() {
	const ttl = time.Second * 300

	ticker := time.NewTicker(time.Second * 20)
	defer ticker.Stop()

	for {
		select {
		case <-g.ctx.Done():
			return

		case <-ticker.C:
			var expiredHost []string

			g.cache.Range(func(key, value interface{}) bool {
				if record, ok := value.(*dnsRecord); ok {
					if time.Since(record.ts) > ttl {
						expiredHost = append(expiredHost, record.host)
					}
				}
				return true
			})

			for _, host := range expiredHost {
				g.cache.Delete(host)
			}
		}
	}
}

func (g *googleDNS) Query(host string) (network string, ip string, err error) {
	// if host is ip4 or ip6
	if netIP := net.ParseIP(host); netIP != nil {
		if netIP.To4() != nil {
			network = "tcp4"
			ip = host

		} else {
			network = "tcp6"
			ip = fmt.Sprintf("[%s]", host)
		}
		return
	}

	var bingo bool
	if bingo, network, ip = g.localQuery(host); bingo {
		return
	}

	client := new(http.Client)
	if g.dummy {
		proxy, _ := url.Parse("http://127.0.0.1:8087")
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
	}

	for _, _type := range []string{"aaaa", "A"} {
		url := fmt.Sprintf("https://dns.google/resolve?name=%s&type=%s&edns_client_subnet=0:0:0::/0&random_padding=", host, _type)
		if padding := 150 - len(url); padding > 0 { // dynamic random_padding, max url len 150
			if padding%2 != 0 {
				padding++
			}

			buf := make([]byte, padding/2)
			io.ReadFull(rand.Reader, buf)
			url += hex.EncodeToString(buf)
		}

		var resp *http.Response
		if resp, err = client.Get(url); err != nil {
			err = errors.WithStack(err)
			return
		}

		var body []byte
		if body, err = ioutil.ReadAll(resp.Body); err != nil {
			err = errors.WithStack(err)
			return
		}
		resp.Body.Close() // close body

		var result struct {
			Answer []struct {
				Data string `json:"data"`
			} `json:"Answer"`
		}

		if err = json.Unmarshal(body, &result); err != nil {
			err = errors.WithStack(err)
			return
		}

		for _, answer := range result.Answer {
			if net.ParseIP(answer.Data) != nil {
				if _type == "aaaa" {
					network = "tcp6"
					ip = fmt.Sprintf("[%s]", answer.Data)

				} else {
					network = "tcp4"
					ip = answer.Data
				}

				g.store(host, network, ip)
				return
			}
		}
	}

	return
}
