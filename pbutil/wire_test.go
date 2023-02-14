package pbutil

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bluekaki/pkg/pbutil/testdata"

	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestWire(t *testing.T) {
	raw := []byte{255, 255, 255, 255, 255, 255, 255, 255, 126, 123, 33, 2, 0, 0, 0, 0, 186, 67, 209, 250, 255, 255, 255, 255, 77, 0, 0, 0, 0, 0, 0, 0, 222, 126, 123, 108, 238, 255, 255, 255}
	t.Log(string(raw))

	runes := []rune(string(raw))
	t.Log(runes)

	t.Log([]rune("hello world"))

	req := &pb.HelloRequest{
		Sequence:  1029,
		Message:   "Hello World !!! ",
		Timestamp: timestamppb.New(time.Now()),
		Duration:  durationpb.New(time.Second * 17),
		Status:    pb.HelloRequest_Shutdown,
		Payloads: []*pb.HelloRequest_Payload{
			{Raw: []byte("abc")},
			{
				Raw: []byte("xyz"),
				Meta: &pb.HelloRequest_Payload_Metadata{
					Ts:    "2022-02-22",
					Nonce: []string{"1", "2", "3", "4", "5"},
				},
			},
		},
		NickName: &pb.HelloRequest_FirstName{FirstName: "jack"},
		Meta: map[int32]bool{
			200: true,
			500: false,
		},
		Nonce: []int64{-1, 35748734, -86948934, 77, -75489378594},
		Ack:   true,
		Memo:  "// TODO ...",
	}

	raw, _ = proto.Marshal(req)

	buf := bytes.NewBuffer(nil)
	parseWire(raw, t, buf)

	t.Log(buf.String())
}

func parseWire(raw []byte, t *testing.T, buf *bytes.Buffer) bool {
	var groups []byte
	rawSize := len(raw)
	for index := 0; index < rawSize; {
		groups, index = readVarint(index, raw)
		offset, _type := key(groups)

		switch _type {
		case 0: // Varint
			groups, index = readVarint(index, raw)
			varint, zigzag, err := parseVarint(groups)
			if err != nil {
				return false
			}

			// t.Log(offset, "varint:", varint, "zigzag:", zigzag)
			buf.WriteString(fmt.Sprintf("[%d] varint: %d, zigzag: %d; \n", offset, varint, zigzag))

		case 1: // 64-bit
			cursor := index + 8
			if cursor > rawSize {
				return false
			}

			lump := raw[index:cursor]
			index = cursor

			bits := binary.LittleEndian.Uint64(lump)
			sfixed64 := int64(bits)
			fixed64 := bits
			double := math.Float64frombits(bits)

			// t.Log(offset, "sfixed64:", sfixed64, "fixed64:", fixed64, "double:", double)
			buf.WriteString(fmt.Sprintf("[%d] sfixed64: %d, fixed64: %d, double: %f; \n", offset, sfixed64, fixed64, double))

		case 5: // 32-bit
			cursor := index + 4
			if cursor > rawSize {
				return false
			}

			lump := raw[index:cursor]
			index = cursor

			bits := binary.LittleEndian.Uint32(lump)
			sfixed32 := int32(bits)
			fixed32 := bits
			float := math.Float32frombits(bits)

			// t.Log(offset, "sfixed32:", sfixed32, "fixed32:", fixed32, "float:", float)
			buf.WriteString(fmt.Sprintf("[%d] sfixed32: %d, fixed32: %d, float: %f; \n", offset, sfixed32, fixed32, float))

		case 2: // Length-delimited
			groups, index = readVarint(index, raw)
			length := positiveVarint(groups)
			if length <= 0 {
				return false
			}

			cursor := index + length
			if cursor > rawSize {
				return false
			}

			payload := raw[index:cursor]
			index = cursor

			if offset == 11 {
				t.Log(payload)
			}

			if length%8 == 0 { // []64-bit
				buffer := bytes.NewBuffer(nil)
				for k := 0; k < length; k += 8 {
					lump := payload[k : k+8]

					bits := binary.LittleEndian.Uint64(lump)
					sfixed64 := int64(bits)
					fixed64 := bits
					double := math.Float64frombits(bits)

					buffer.WriteString(fmt.Sprintf("[64] sfixed64: %d, fixed64: %d, double: %f; \n", sfixed64, fixed64, double))
				}

				if offset == 2000 {
					t.Log("*******", buffer.String())
				}
			}

			if length%4 == 0 { // []32-bit
				buffer := bytes.NewBuffer(nil)
				for k := 0; k < length; k += 4 {
					lump := payload[k : k+4]

					bits := binary.LittleEndian.Uint32(lump)
					sfixed32 := int32(bits)
					fixed32 := bits
					float := math.Float32frombits(bits)

					buffer.WriteString(fmt.Sprintf("[32] sfixed32: %d, fixed32: %d, float: %f; \n", sfixed32, fixed32, float))
				}

				if offset == 2000 {
					t.Log("*******", buffer.String())
				}
			}

			{ // []varint
				buffer := bytes.NewBuffer(nil)
				for k := 0; k < length; {
					groups, k = readVarint(k, payload)
					varint, zigzag, err := parseVarint(groups)
					if err != nil {
						goto loop
					}

					buffer.WriteString(fmt.Sprintf("[var] varint: %d, zigzag: %d; \n", varint, zigzag))
				}

				if offset == 2000 {
					t.Log("*******", buffer.String())
				}
			}

		loop:
			buffer := bytes.NewBuffer(nil)
			if parseWire(payload, t, buffer) {
				buf.WriteString(fmt.Sprintf("[%d] { \n", offset))
				buf.Write(buffer.Bytes())
				buf.WriteString("}; \n")

			} else {
				buf.WriteString(fmt.Sprintf("[%d] %s; \n", offset, string(payload)))
			}

		default:
			return false
		}
	}

	return true
}

func readVarint(index int, raw []byte) ([]byte, int) {
	var groups []byte // 7bits
	for ; index < len(raw); index++ {
		if char := raw[index]; char>>7 == 1 {
			groups = append(groups, char<<1>>1) // remove the first mark bit
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
	for k := len(groups) - 1; k >= 0; k-- { // reverse 7bit-groups
		tmp = append(tmp, fmt.Sprintf("%07b", groups[k]))
	}

	bits := strings.Join(tmp, "")

	offset, _ = strconv.ParseUint(bits[:len(bits)-3], 2, 32)
	_type, _ = strconv.ParseUint(bits[len(bits)-3:], 2, 3) // last 3 bits as data type
	return
}

func parseVarint(groups []byte) (_varint, _zigzag interface{}, err error) {
	bits := make([]string, 0, len(groups))
	for k := len(groups) - 1; k >= 0; k-- { // reverse 7bit-groups
		bits = append(bits, fmt.Sprintf("%07b", groups[k]))
	}

	var unsigned uint64
	if unsigned, err = strconv.ParseUint(strings.Join(bits, ""), 2, 64); err != nil {
		return
	}
	_zigzag = int64(unsigned>>1) ^ (-int64(unsigned & 1))

	if len(groups) == 10 { // negative int32/int64; special for big sint64.
		_varint = int64(unsigned)

	} else {
		_varint = unsigned
	}

	return
}

func positiveVarint(groups []byte) int {
	bits := make([]string, 0, len(groups))
	for k := len(groups) - 1; k >= 0; k-- { // reverse 7bit-groups
		bits = append(bits, fmt.Sprintf("%07b", groups[k]))
	}

	unsigned, _ := strconv.ParseUint(strings.Join(bits, ""), 2, 64)
	return int(unsigned)
}
