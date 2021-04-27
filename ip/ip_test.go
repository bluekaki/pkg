package ip

import (
	"encoding/binary"
	"math/big"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	zoneUS4 *Zone
	zoneCA4 *Zone
	zoneAU4 *Zone

	zoneUS16 *Zone
	zoneCA16 *Zone
	zoneAU16 *Zone
)

func Test_Init(t *testing.T) {
	assert := assert.New(t)
	var err error

	zoneUS4, err = MkZone("us", "http://ipverse.net/ipblocks/data/countries/us.zone")
	assert.Nil(err)

	zoneCA4, err = MkZone("ca", "http://ipverse.net/ipblocks/data/countries/ca.zone")
	assert.Nil(err)

	zoneAU4, err = MkZone("au", "http://ipverse.net/ipblocks/data/countries/au.zone")
	assert.Nil(err)

	zoneUS16, err = MkZone("us", "http://ipverse.net/ipblocks/data/countries/us-ipv6.zone")
	assert.Nil(err)

	zoneCA16, err = MkZone("ca", "http://ipverse.net/ipblocks/data/countries/ca-ipv6.zone")
	assert.Nil(err)

	zoneAU16, err = MkZone("au", "http://ipverse.net/ipblocks/data/countries/au-ipv6.zone")
	assert.Nil(err)
}

func Test_IP4(t *testing.T) {
	assert := assert.New(t)
	filter, err := NewFilter(zoneUS4, zoneCA4)
	assert.Nil(err)

	// us bingo
	// us: 23.19.0.0/19  23.19.0.0 - 23.19.31.255
	first := binary.BigEndian.Uint32([]byte{23, 19, 0, 0})
	last := binary.BigEndian.Uint32([]byte{23, 19, 31, 255})
	for k := first; k <= last; k++ {
		ip := make([]byte, 4)
		binary.BigEndian.PutUint32(ip, k)
		ok, name, err := filter.Bingo(net.IP(ip).String())
		assert.Nil(err)
		assert.True(ok)
		assert.Equal(name, "us")
	}

	// ca bingo
	// ca: 23.254.0.0/17  23.254.0.0 - 23.254.127.255
	first = binary.BigEndian.Uint32([]byte{23, 254, 0, 0})
	last = binary.BigEndian.Uint32([]byte{23, 254, 127, 255})
	for k := first; k <= last; k++ {
		ip := make([]byte, 4)
		binary.BigEndian.PutUint32(ip, k)
		ok, name, err := filter.Bingo(net.IP(ip).String())
		assert.Nil(err)
		assert.True(ok)
		assert.Equal(name, "ca")
	}

	// au not bingo
	// au: 1.178.0.0/16  1.178.0.0 - 1.178.255.255
	first = binary.BigEndian.Uint32([]byte{1, 178, 0, 0})
	last = binary.BigEndian.Uint32([]byte{1, 178, 255, 255})
	for k := first; k <= last; k++ {
		ip := make([]byte, 4)
		binary.BigEndian.PutUint32(ip, k)
		ok, name, err := filter.Bingo(net.IP(ip).String())
		assert.Nil(err)
		assert.False(ok)
		assert.Empty(name)
	}
}

func Test_IP16(t *testing.T) {
	assert := assert.New(t)
	filter, err := NewFilter(zoneUS16, zoneCA16)
	assert.Nil(err)

	// us bingo
	// us: 2600:800::/27
	raw := big.NewInt(0).SetBytes(net.ParseIP("2600:800::"))
	for k := 1; k <= 10000; k++ {
		ok, name, err := filter.Bingo(net.IP(raw.Add(raw, big.NewInt(1)).Bytes()).String())
		assert.Nil(err)
		assert.True(ok)
		assert.Equal(name, "us")
	}

	// ca bingo
	// ca: 2001:568::/29
	raw = big.NewInt(0).SetBytes(net.ParseIP("2001:568::"))
	for k := 1; k <= 10000; k++ {
		ok, name, err := filter.Bingo(net.IP(raw.Add(raw, big.NewInt(1)).Bytes()).String())
		assert.Nil(err)
		assert.True(ok)
		assert.Equal(name, "ca")
	}

	// au not bingo
	// au: 2001:df0:2aa::/48
	raw = big.NewInt(0).SetBytes(net.ParseIP("2001:df0:2aa::"))
	for k := 1; k <= 10000; k++ {
		ok, name, err := filter.Bingo(net.IP(raw.Add(raw, big.NewInt(1)).Bytes()).String())
		assert.Nil(err)
		assert.False(ok)
		assert.Empty(name)
	}
}
