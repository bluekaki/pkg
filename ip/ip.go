package ip

import (
	"encoding/binary"
	"math"
	"net"
	"sort"

	"github.com/bluekaki/pkg/errors"
)

const intervalSize = 1024

var _ Filter = (*filter)(nil)

// Filter support operations of ip4/ip16 filter
type Filter interface {
	// Bingo check ip likes 23.19.0.1 or 2600:800:: wheter in filter
	Bingo(ip string) (ok bool, name string, err error)
}

type interval4 struct {
	zone string
	min  uint32
	max  uint32
}

type interval16 struct {
	zone string
	min  [2]uint64
	max  [2]uint64
}

type block4 struct {
	min       uint32
	max       uint32
	intervals []*interval4
}

type block16 struct {
	min       [2]uint64
	max       [2]uint64
	intervals []*interval16
}

type filter struct {
	ip4  []*block4
	ip16 []*block16
}

// Zone define cidr(s)
type Zone struct {
	Name string
	CIDR []string
}

// NewFilter return new instance
// filter no support dynamic change zones reason for performance;
// if zones changed, just new another filter.
func NewFilter(zones ...*Zone) (Filter, error) {
	f := new(filter)

	if len(zones) == 0 {
		return nil, errors.New("zones required")
	}

	var interval4s []*interval4
	var interval16s []*interval16

	for _, zone := range zones {
		for _, cidr := range zone.CIDR {
			_, netip, err := net.ParseCIDR(cidr)
			if err != nil {
				return nil, errors.Wrapf(err, "parse cidr %s err", cidr)
			}

			ones, bits := netip.Mask.Size()
			switch bits {
			case 32:
				raw := binary.BigEndian.Uint32(netip.IP)
				min := raw
				max := min | math.MaxUint32>>ones

				interval4s = append(interval4s, &interval4{
					zone: zone.Name,
					min:  min,
					max:  max,
				})

			case 128:
				min, max := f.shift(netip.IP, ones)

				interval16s = append(interval16s, &interval16{
					zone: zone.Name,
					min:  min,
					max:  max,
				})
			}
		}
	}

	if len(interval4s) == 0 && len(interval16s) == 0 {
		return nil, errors.New("both ip4 and ip16 are empty")
	}

	f.initIP4(interval4s)
	f.initIP16(interval16s)

	return f, nil
}

func (f *filter) shift(ip16 net.IP, ones int) (min, max [2]uint64) {
	bits := make([]uint8, 128)
	for i, v := range ip16 {
		for k := 0; k < 8; k++ {
			bits[i*8+k] = v >> (7 - k) & 1
		}
	}

	toBytes := func(bits []uint8) []byte {
		bytes := make([]byte, 16)
		for i := range bytes {
			for k := 0; k < 8; k++ {
				bytes[i] |= bits[i*8+k] << (7 - k)
			}
		}
		return bytes
	}

	min[0] = binary.BigEndian.Uint64(ip16[:8])
	min[1] = binary.BigEndian.Uint64(ip16[8:])

	for k := ones; k < 128; k++ {
		bits[k] = 1
	}

	bytes := toBytes(bits)
	max[0] = binary.BigEndian.Uint64(bytes[:8])
	max[1] = binary.BigEndian.Uint64(bytes[8:])

	return
}

func (f *filter) initIP4(intervals []*interval4) {
	if intervals == nil {
		return
	}

	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i].max < intervals[j].max
	})

	blockSize := len(intervals) / intervalSize
	if len(intervals)%intervalSize != 0 {
		blockSize++
	}
	f.ip4 = make([]*block4, blockSize)

	for k := 0; k < blockSize-1; k++ {
		f.ip4[k] = &block4{
			min:       intervals[k*intervalSize].min,
			max:       intervals[(k+1)*intervalSize-1].max,
			intervals: intervals[k*intervalSize : (k+1)*intervalSize],
		}
	}
	f.ip4[blockSize-1] = &block4{
		min:       intervals[(blockSize-1)*intervalSize].min,
		max:       intervals[len(intervals)-1].max,
		intervals: intervals[(blockSize-1)*intervalSize:],
	}
}

func (f *filter) initIP16(intervals []*interval16) {
	if intervals == nil {
		return
	}

	sort.Slice(intervals, func(i, j int) bool {
		return f.ip16Less(intervals[i].max, intervals[j].max)
	})

	blockSize := len(intervals) / intervalSize
	if len(intervals)%intervalSize != 0 {
		blockSize++
	}
	f.ip16 = make([]*block16, blockSize)

	for k := 0; k < blockSize-1; k++ {
		f.ip16[k] = &block16{
			min:       intervals[k*intervalSize].min,
			max:       intervals[(k+1)*intervalSize-1].max,
			intervals: intervals[k*intervalSize : (k+1)*intervalSize],
		}
	}
	f.ip16[blockSize-1] = &block16{
		min:       intervals[(blockSize-1)*intervalSize].min,
		max:       intervals[len(intervals)-1].max,
		intervals: intervals[(blockSize-1)*intervalSize:],
	}
}

type result int

const (
	equal   result = 0
	greater result = 1
	less    result = -1
)

func (f *filter) compIP16(x, y [2]uint64) result {
	for i := 0; i < 2; i++ {
		if x[i] < y[i] {
			return less

		} else if x[i] > y[i] {
			return greater
		}
	}
	return equal
}

func (f *filter) ip16Less(x, y [2]uint64) bool {
	return f.compIP16(x, y) == less
}

func (f *filter) ip16LessEqual(x, y [2]uint64) bool {
	result := f.compIP16(x, y)
	return result == less || result == equal
}

func (f *filter) ip16Greater(x, y [2]uint64) bool {
	return f.compIP16(x, y) == greater
}

func (f *filter) ip16GreaterEqual(x, y [2]uint64) bool {
	result := f.compIP16(x, y)
	return result == greater || result == equal
}

func (f *filter) Bingo(ip string) (ok bool, zone string, err error) {
	netIP := net.ParseIP(ip)
	if netIP == nil {
		err = errors.Errorf("%s is not ip4 or ip16", ip)
		return
	}

	if ip := []byte(netIP.To4()); ip != nil {
		ok, zone = f.searchIP4(ip)
	} else {
		ok, zone = f.searchIP16(netIP)
	}

	return
}

func (f *filter) searchIP4(ip net.IP) (ok bool, zone string) {
	raw := binary.BigEndian.Uint32(ip)

	index := sort.Search(len(f.ip4), func(i int) bool {
		return raw <= f.ip4[i].max
	})
	if index != -1 && index < len(f.ip4) && f.ip4[index].min <= raw {
		intervals := f.ip4[index].intervals

		index = sort.Search(len(intervals), func(i int) bool {
			return raw <= intervals[i].max
		})
		if index != -1 && index < len(intervals) && intervals[index].min <= raw {
			ok, zone = true, intervals[index].zone
		}
	}

	return
}

func (f *filter) searchIP16(ip net.IP) (ok bool, zone string) {
	var raw [2]uint64
	raw[0] = binary.BigEndian.Uint64(ip[:8])
	raw[1] = binary.BigEndian.Uint64(ip[8:])

	index := sort.Search(len(f.ip16), func(i int) bool {
		return f.ip16LessEqual(raw, f.ip16[i].max)
	})
	if index != -1 && index < len(f.ip16) && f.ip16LessEqual(f.ip16[index].min, raw) {
		intervals := f.ip16[index].intervals

		index = sort.Search(len(intervals), func(i int) bool {
			return f.ip16LessEqual(raw, intervals[i].max)
		})
		if index != -1 && index < len(intervals) && f.ip16LessEqual(intervals[index].min, raw) {
			ok, zone = true, intervals[index].zone
		}
	}

	return
}
