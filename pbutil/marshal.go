package pbutil

import (
	"bytes"
	"encoding/json"

	"github.com/bluekaki/pkg/errors"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

var jsonMarshal = &jsonpb.Marshaler{
	OrigName:     true,
	EmitDefaults: true,
}

var jsonUnmarshaler = &jsonpb.Unmarshaler{
	AllowUnknownFields: true,
}

// ProtoMessage2JSON marshal protobuf message to json message
func ProtoMessage2JSON(message proto.Message) (json.RawMessage, error) {
	if message == nil {
		return nil, errors.New("message required")
	}

	buf := bytes.NewBuffer(nil)
	err := jsonMarshal.Marshal(buf, message)
	if err != nil {
		return nil, errors.Wrap(err, "marshal protobuf message to json err")
	}

	return json.RawMessage(buf.Bytes()), nil
}

// JSON2ProtoMessage unmarshal json to protobuf message
func JSON2ProtoMessage(raw []byte, message proto.Message) error {
	if err := jsonUnmarshaler.Unmarshal(bytes.NewReader(raw), message); err != nil {
		return errors.Wrap(err, "unmarshal json to protobuf message err")
	}

	return nil
}
