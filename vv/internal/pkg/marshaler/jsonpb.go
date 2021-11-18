package marshaler

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

var jsonPbMarshaler = &runtime.JSONPb{
	MarshalOptions: protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	},
	UnmarshalOptions: protojson.UnmarshalOptions{
		DiscardUnknown: true,
	},
}

func NewJSONPbMarshaler() runtime.Marshaler {
	return &runtime.HTTPBodyMarshaler{Marshaler: jsonPbMarshaler}
}
