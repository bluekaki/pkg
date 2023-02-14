package dns

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDNS(t *testing.T) {
	assert := assert.New(t)

	dns := NewGoogleDNS(WithTunnel())
	defer dns.Close()

	network, ip, err := dns.Query("127.0.0.1")
	t.Log(network)
	t.Log(ip)

	assert.Nil(err)
	assert.Equal(network, "tcp4")
	assert.Equal(ip, "127.0.0.1")

	network, ip, err = dns.Query("0:0:0::")
	t.Log(network)
	t.Log(ip)

	assert.Nil(err)
	assert.Equal(network, "tcp6")
	assert.Equal(ip, "[0:0:0::]")

	network, ip, err = dns.Query("google.com")
	t.Log(network)
	t.Log(ip)

	assert.Nil(err)
	assert.Equal(network, "tcp6")
}
