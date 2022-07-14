package pbutil

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/bluekaki/pkg/pbutil/testdata"

	"github.com/golang/protobuf/proto"
	// "google.golang.org/protobuf/types/known/durationpb"
	// "google.golang.org/protobuf/types/known/timestamppb"
)

func TestWire(t *testing.T) {
	// req := &pb.HelloRequest{
	// 	Sequence:  1029,
	// 	Message:   "Hello World !!! 你好，世界。",
	// 	Timestamp: timestamppb.New(time.Now()),
	// 	Duration:  durationpb.New(time.Second * 17),
	// 	Status:    pb.HelloRequest_RUNNING,
	// 	Payloads: []*pb.HelloRequest_Payload{
	// 		{Raw: []byte("abc")},
	// 		{Raw: []byte("xyz")},
	// 	},
	// 	NickName: &pb.HelloRequest_FirstName{FirstName: "jack"},
	// 	Meta: map[int32]bool{
	// 		200: true,
	// 		500: false,
	// 	},
	// }

	req := &pb.Numbers{
		INT32:  102400,
		UINT32: 204800,
		INT64:  409600,
		UINT64: 819200,
	}

	raw, _ := proto.Marshal(req)
	t.Log(raw)

	// var t0, t1 uint
	var groups []byte
	for index := 0; index < len(raw); {
		groups, index = readVarint(index, raw)
		// t.Log(groups)

		offset, _type := key(groups)
		switch _type {
		case 0: // Varint
			groups, index = readVarint(index, raw)
			t.Log(offset, varint(groups))
		}
	}

	// t.Log("01:", raw[0]>>3, raw[0]<<5>>5)
	// x, y := raw[1], raw[2]
	// t.Log(x>>7, y>>7)

	// xy := fmt.Sprintf("%07b%07b", y<<1>>1, x<<1>>1)
	// t.Log(strconv.ParseUint(xy, 2, 16))
}

func readVarint(index int, raw []byte) ([]byte, int) {
	var groups []byte // 7bits
	for ; index < len(raw); index++ {
		if char := raw[index]; char>>7 == 1 {
			groups = append(groups, char<<1>>1) // remove the first bit of mark
			continue
		}
		groups = append(groups, raw[index]<<1>>1)

		index++ // the original point for next loop
		break
	}

	return groups, index
}

func key(groups []byte) (offset uint64, _type uint64) {
	tmp := make([]string, 0, len(groups))
	for k := len(groups) - 1; k >= 0; k-- { // reverse group with 7bits
		tmp = append(tmp, fmt.Sprintf("%07b", groups[k]))
	}

	bits := strings.Join(tmp, "")

	offset, _ = strconv.ParseUint(bits[:len(bits)-3], 2, 32)
	_type, _ = strconv.ParseUint(bits[len(bits)-3:], 2, 3) // last 3 bits as data type
	return
}

func varint(groups []byte) uint64 {
	tmp := make([]string, 0, len(groups))
	for k := len(groups) - 1; k >= 0; k-- { // reverse group with 7bits
		tmp = append(tmp, fmt.Sprintf("%07b", groups[k]))
	}

	bits := strings.Join(tmp, "")
	val, _ := strconv.ParseUint(bits, 2, 64)
	return val
}
