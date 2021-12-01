package marshaler

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

type jsonPb struct {
	runtime.JSONPb
}

func (j *jsonPb) ContentType(_ interface{}) string {
	return "application/json; charset=utf-8"
}

var jsonPbMarshaler *jsonPb

func init() {
	jsonPbMarshaler = new(jsonPb)
	jsonPbMarshaler.MarshalOptions = protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}

	jsonPbMarshaler.UnmarshalOptions = protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
}

func NewJSONPbMarshaler() runtime.Marshaler {
	return &runtime.HTTPBodyMarshaler{Marshaler: jsonPbMarshaler}
}
