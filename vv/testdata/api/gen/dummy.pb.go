// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.18.1
// source: dummy.proto

package dummy

import (
	_ "github.com/bluekaki/pkg/vv/pkg/plugin/interceptor/options"
	_ "github.com/bluekaki/pkg/vv/pkg/plugin/protoc-gen-message-validator/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type EchoReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *EchoReq) Reset() {
	*x = EchoReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dummy_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EchoReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EchoReq) ProtoMessage() {}

func (x *EchoReq) ProtoReflect() protoreflect.Message {
	mi := &file_dummy_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EchoReq.ProtoReflect.Descriptor instead.
func (*EchoReq) Descriptor() ([]byte, []int) {
	return file_dummy_proto_rawDescGZIP(), []int{0}
}

func (x *EchoReq) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type EchoResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	Ack     bool   `protobuf:"varint,2,opt,name=ack,proto3" json:"ack,omitempty"`
}

func (x *EchoResp) Reset() {
	*x = EchoResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dummy_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EchoResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EchoResp) ProtoMessage() {}

func (x *EchoResp) ProtoReflect() protoreflect.Message {
	mi := &file_dummy_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EchoResp.ProtoReflect.Descriptor instead.
func (*EchoResp) Descriptor() ([]byte, []int) {
	return file_dummy_proto_rawDescGZIP(), []int{1}
}

func (x *EchoResp) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *EchoResp) GetAck() bool {
	if x != nil {
		return x.Ack
	}
	return false
}

var File_dummy_proto protoreflect.FileDescriptor

var file_dummy_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x64, 0x75, 0x6d, 0x6d, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x64,
	0x75, 0x6d, 0x6d, 0x79, 0x1a, 0x19, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x63, 0x65, 0x70, 0x74, 0x6f,
	0x72, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x2d, 0x0a, 0x07, 0x45, 0x63, 0x68, 0x6f, 0x52, 0x65,
	0x71, 0x12, 0x22, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x08, 0xc2, 0xb6, 0x24, 0x04, 0x08, 0x01, 0x28, 0x1e, 0x52, 0x07, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x36, 0x0a, 0x08, 0x45, 0x63, 0x68, 0x6f, 0x52, 0x65, 0x73,
	0x70, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x61,
	0x63, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x03, 0x61, 0x63, 0x6b, 0x32, 0xf1, 0x01,
	0x0a, 0x0c, 0x44, 0x75, 0x6d, 0x6d, 0x79, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x67,
	0x0a, 0x04, 0x45, 0x63, 0x68, 0x6f, 0x12, 0x0e, 0x2e, 0x64, 0x75, 0x6d, 0x6d, 0x79, 0x2e, 0x45,
	0x63, 0x68, 0x6f, 0x52, 0x65, 0x71, 0x1a, 0x0f, 0x2e, 0x64, 0x75, 0x6d, 0x6d, 0x79, 0x2e, 0x45,
	0x63, 0x68, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x22, 0x3e, 0xa2, 0xb0, 0x24, 0x27, 0x0a, 0x09, 0x64,
	0x75, 0x6d, 0x6d, 0x79, 0x5f, 0x73, 0x73, 0x6f, 0x12, 0x0a, 0x64, 0x75, 0x6d, 0x6d, 0x79, 0x5f,
	0x73, 0x69, 0x67, 0x6e, 0x1a, 0x0c, 0x64, 0x75, 0x6d, 0x6d, 0x79, 0x5f, 0x69, 0x70, 0x6c, 0x69,
	0x73, 0x74, 0x20, 0x01, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0d, 0x12, 0x0b, 0x2f, 0x64, 0x75, 0x6d,
	0x6d, 0x79, 0x2f, 0x65, 0x63, 0x68, 0x6f, 0x12, 0x78, 0x0a, 0x0a, 0x53, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x45, 0x63, 0x68, 0x6f, 0x12, 0x0e, 0x2e, 0x64, 0x75, 0x6d, 0x6d, 0x79, 0x2e, 0x45, 0x63,
	0x68, 0x6f, 0x52, 0x65, 0x71, 0x1a, 0x0f, 0x2e, 0x64, 0x75, 0x6d, 0x6d, 0x79, 0x2e, 0x45, 0x63,
	0x68, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x22, 0x45, 0xa2, 0xb0, 0x24, 0x27, 0x0a, 0x09, 0x64, 0x75,
	0x6d, 0x6d, 0x79, 0x5f, 0x73, 0x73, 0x6f, 0x12, 0x0a, 0x64, 0x75, 0x6d, 0x6d, 0x79, 0x5f, 0x73,
	0x69, 0x67, 0x6e, 0x1a, 0x0c, 0x64, 0x75, 0x6d, 0x6d, 0x79, 0x5f, 0x69, 0x70, 0x6c, 0x69, 0x73,
	0x74, 0x20, 0x01, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x14, 0x12, 0x12, 0x2f, 0x64, 0x75, 0x6d, 0x6d,
	0x79, 0x2f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2f, 0x65, 0x63, 0x68, 0x6f, 0x28, 0x01, 0x30,
	0x01, 0x42, 0x0a, 0x5a, 0x08, 0x2e, 0x2f, 0x3b, 0x64, 0x75, 0x6d, 0x6d, 0x79, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_dummy_proto_rawDescOnce sync.Once
	file_dummy_proto_rawDescData = file_dummy_proto_rawDesc
)

func file_dummy_proto_rawDescGZIP() []byte {
	file_dummy_proto_rawDescOnce.Do(func() {
		file_dummy_proto_rawDescData = protoimpl.X.CompressGZIP(file_dummy_proto_rawDescData)
	})
	return file_dummy_proto_rawDescData
}

var file_dummy_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_dummy_proto_goTypes = []interface{}{
	(*EchoReq)(nil),  // 0: dummy.EchoReq
	(*EchoResp)(nil), // 1: dummy.EchoResp
}
var file_dummy_proto_depIdxs = []int32{
	0, // 0: dummy.DummyService.Echo:input_type -> dummy.EchoReq
	0, // 1: dummy.DummyService.StreamEcho:input_type -> dummy.EchoReq
	1, // 2: dummy.DummyService.Echo:output_type -> dummy.EchoResp
	1, // 3: dummy.DummyService.StreamEcho:output_type -> dummy.EchoResp
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_dummy_proto_init() }
func file_dummy_proto_init() {
	if File_dummy_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_dummy_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EchoReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_dummy_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EchoResp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_dummy_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_dummy_proto_goTypes,
		DependencyIndexes: file_dummy_proto_depIdxs,
		MessageInfos:      file_dummy_proto_msgTypes,
	}.Build()
	File_dummy_proto = out.File
	file_dummy_proto_rawDesc = nil
	file_dummy_proto_goTypes = nil
	file_dummy_proto_depIdxs = nil
}
