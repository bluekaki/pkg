// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.18.1
// source: validator_options.proto

package options

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type FieldValidator struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Require     *bool   `protobuf:"varint,1,opt,name=require,proto3,oneof" json:"require,omitempty"`
	Eq          *string `protobuf:"bytes,2,opt,name=eq,proto3,oneof" json:"eq,omitempty"`                                        // equal to
	Ne          *string `protobuf:"bytes,3,opt,name=ne,proto3,oneof" json:"ne,omitempty"`                                        // not equal to
	Lt          *uint32 `protobuf:"varint,4,opt,name=lt,proto3,oneof" json:"lt,omitempty"`                                       // less then
	Le          *uint32 `protobuf:"varint,5,opt,name=le,proto3,oneof" json:"le,omitempty"`                                       // less than or equal to
	Gt          *uint32 `protobuf:"varint,6,opt,name=gt,proto3,oneof" json:"gt,omitempty"`                                       // greater than
	Ge          *uint32 `protobuf:"varint,7,opt,name=ge,proto3,oneof" json:"ge,omitempty"`                                       // greater than or equal to
	MaxCap      *uint32 `protobuf:"varint,8,opt,name=max_cap,json=maxCap,proto3,oneof" json:"max_cap,omitempty"`                 // capacity by slice/map/bytes
	MinCap      *uint32 `protobuf:"varint,9,opt,name=min_cap,json=minCap,proto3,oneof" json:"min_cap,omitempty"`                 // capacity by slice/map/bytes
	CstDatetime *bool   `protobuf:"varint,10,opt,name=cst_datetime,json=cstDatetime,proto3,oneof" json:"cst_datetime,omitempty"` // 2006-01-02 15:04:05
	CstMinute   *bool   `protobuf:"varint,11,opt,name=cst_minute,json=cstMinute,proto3,oneof" json:"cst_minute,omitempty"`       // 2006-01-02 15:04
	CnMobile    *bool   `protobuf:"varint,12,opt,name=cn_mobile,json=cnMobile,proto3,oneof" json:"cn_mobile,omitempty"`          // 10000000000
	CstDay      *bool   `protobuf:"varint,13,opt,name=cst_day,json=cstDay,proto3,oneof" json:"cst_day,omitempty"`                // 2006-01-02
	Duration    *bool   `protobuf:"varint,14,opt,name=duration,proto3,oneof" json:"duration,omitempty"`                          // 7s
}

func (x *FieldValidator) Reset() {
	*x = FieldValidator{}
	if protoimpl.UnsafeEnabled {
		mi := &file_validator_options_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FieldValidator) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FieldValidator) ProtoMessage() {}

func (x *FieldValidator) ProtoReflect() protoreflect.Message {
	mi := &file_validator_options_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FieldValidator.ProtoReflect.Descriptor instead.
func (*FieldValidator) Descriptor() ([]byte, []int) {
	return file_validator_options_proto_rawDescGZIP(), []int{0}
}

func (x *FieldValidator) GetRequire() bool {
	if x != nil && x.Require != nil {
		return *x.Require
	}
	return false
}

func (x *FieldValidator) GetEq() string {
	if x != nil && x.Eq != nil {
		return *x.Eq
	}
	return ""
}

func (x *FieldValidator) GetNe() string {
	if x != nil && x.Ne != nil {
		return *x.Ne
	}
	return ""
}

func (x *FieldValidator) GetLt() uint32 {
	if x != nil && x.Lt != nil {
		return *x.Lt
	}
	return 0
}

func (x *FieldValidator) GetLe() uint32 {
	if x != nil && x.Le != nil {
		return *x.Le
	}
	return 0
}

func (x *FieldValidator) GetGt() uint32 {
	if x != nil && x.Gt != nil {
		return *x.Gt
	}
	return 0
}

func (x *FieldValidator) GetGe() uint32 {
	if x != nil && x.Ge != nil {
		return *x.Ge
	}
	return 0
}

func (x *FieldValidator) GetMaxCap() uint32 {
	if x != nil && x.MaxCap != nil {
		return *x.MaxCap
	}
	return 0
}

func (x *FieldValidator) GetMinCap() uint32 {
	if x != nil && x.MinCap != nil {
		return *x.MinCap
	}
	return 0
}

func (x *FieldValidator) GetCstDatetime() bool {
	if x != nil && x.CstDatetime != nil {
		return *x.CstDatetime
	}
	return false
}

func (x *FieldValidator) GetCstMinute() bool {
	if x != nil && x.CstMinute != nil {
		return *x.CstMinute
	}
	return false
}

func (x *FieldValidator) GetCnMobile() bool {
	if x != nil && x.CnMobile != nil {
		return *x.CnMobile
	}
	return false
}

func (x *FieldValidator) GetCstDay() bool {
	if x != nil && x.CstDay != nil {
		return *x.CstDay
	}
	return false
}

func (x *FieldValidator) GetDuration() bool {
	if x != nil && x.Duration != nil {
		return *x.Duration
	}
	return false
}

type MediaValidator struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ContentType *string `protobuf:"bytes,1,opt,name=content_type,json=contentType,proto3,oneof" json:"content_type,omitempty"`
}

func (x *MediaValidator) Reset() {
	*x = MediaValidator{}
	if protoimpl.UnsafeEnabled {
		mi := &file_validator_options_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MediaValidator) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MediaValidator) ProtoMessage() {}

func (x *MediaValidator) ProtoReflect() protoreflect.Message {
	mi := &file_validator_options_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MediaValidator.ProtoReflect.Descriptor instead.
func (*MediaValidator) Descriptor() ([]byte, []int) {
	return file_validator_options_proto_rawDescGZIP(), []int{1}
}

func (x *MediaValidator) GetContentType() string {
	if x != nil && x.ContentType != nil {
		return *x.ContentType
	}
	return ""
}

var file_validator_options_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*FieldValidator)(nil),
		Field:         74600,
		Name:          "validator.field",
		Tag:           "bytes,74600,opt,name=field",
		Filename:      "validator_options.proto",
	},
	{
		ExtendedType:  (*descriptorpb.MessageOptions)(nil),
		ExtensionType: (*MediaValidator)(nil),
		Field:         74601,
		Name:          "validator.media",
		Tag:           "bytes,74601,opt,name=media",
		Filename:      "validator_options.proto",
	},
}

// Extension fields to descriptorpb.FieldOptions.
var (
	// optional validator.FieldValidator field = 74600;
	E_Field = &file_validator_options_proto_extTypes[0]
)

// Extension fields to descriptorpb.MessageOptions.
var (
	// optional validator.MediaValidator media = 74601;
	E_Media = &file_validator_options_proto_extTypes[1]
)

var File_validator_options_proto protoreflect.FileDescriptor

var file_validator_options_proto_rawDesc = []byte{
	0x0a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x6f, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x76, 0x61, 0x6c, 0x69, 0x64,
	0x61, 0x74, 0x6f, 0x72, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xab, 0x04, 0x0a, 0x0e, 0x46, 0x69, 0x65, 0x6c, 0x64,
	0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x12, 0x1d, 0x0a, 0x07, 0x72, 0x65, 0x71,
	0x75, 0x69, 0x72, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x48, 0x00, 0x52, 0x07, 0x72, 0x65,
	0x71, 0x75, 0x69, 0x72, 0x65, 0x88, 0x01, 0x01, 0x12, 0x13, 0x0a, 0x02, 0x65, 0x71, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x48, 0x01, 0x52, 0x02, 0x65, 0x71, 0x88, 0x01, 0x01, 0x12, 0x13, 0x0a,
	0x02, 0x6e, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x02, 0x52, 0x02, 0x6e, 0x65, 0x88,
	0x01, 0x01, 0x12, 0x13, 0x0a, 0x02, 0x6c, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x48, 0x03,
	0x52, 0x02, 0x6c, 0x74, 0x88, 0x01, 0x01, 0x12, 0x13, 0x0a, 0x02, 0x6c, 0x65, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x0d, 0x48, 0x04, 0x52, 0x02, 0x6c, 0x65, 0x88, 0x01, 0x01, 0x12, 0x13, 0x0a, 0x02,
	0x67, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0d, 0x48, 0x05, 0x52, 0x02, 0x67, 0x74, 0x88, 0x01,
	0x01, 0x12, 0x13, 0x0a, 0x02, 0x67, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0d, 0x48, 0x06, 0x52,
	0x02, 0x67, 0x65, 0x88, 0x01, 0x01, 0x12, 0x1c, 0x0a, 0x07, 0x6d, 0x61, 0x78, 0x5f, 0x63, 0x61,
	0x70, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0d, 0x48, 0x07, 0x52, 0x06, 0x6d, 0x61, 0x78, 0x43, 0x61,
	0x70, 0x88, 0x01, 0x01, 0x12, 0x1c, 0x0a, 0x07, 0x6d, 0x69, 0x6e, 0x5f, 0x63, 0x61, 0x70, 0x18,
	0x09, 0x20, 0x01, 0x28, 0x0d, 0x48, 0x08, 0x52, 0x06, 0x6d, 0x69, 0x6e, 0x43, 0x61, 0x70, 0x88,
	0x01, 0x01, 0x12, 0x26, 0x0a, 0x0c, 0x63, 0x73, 0x74, 0x5f, 0x64, 0x61, 0x74, 0x65, 0x74, 0x69,
	0x6d, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x08, 0x48, 0x09, 0x52, 0x0b, 0x63, 0x73, 0x74, 0x44,
	0x61, 0x74, 0x65, 0x74, 0x69, 0x6d, 0x65, 0x88, 0x01, 0x01, 0x12, 0x22, 0x0a, 0x0a, 0x63, 0x73,
	0x74, 0x5f, 0x6d, 0x69, 0x6e, 0x75, 0x74, 0x65, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x08, 0x48, 0x0a,
	0x52, 0x09, 0x63, 0x73, 0x74, 0x4d, 0x69, 0x6e, 0x75, 0x74, 0x65, 0x88, 0x01, 0x01, 0x12, 0x20,
	0x0a, 0x09, 0x63, 0x6e, 0x5f, 0x6d, 0x6f, 0x62, 0x69, 0x6c, 0x65, 0x18, 0x0c, 0x20, 0x01, 0x28,
	0x08, 0x48, 0x0b, 0x52, 0x08, 0x63, 0x6e, 0x4d, 0x6f, 0x62, 0x69, 0x6c, 0x65, 0x88, 0x01, 0x01,
	0x12, 0x1c, 0x0a, 0x07, 0x63, 0x73, 0x74, 0x5f, 0x64, 0x61, 0x79, 0x18, 0x0d, 0x20, 0x01, 0x28,
	0x08, 0x48, 0x0c, 0x52, 0x06, 0x63, 0x73, 0x74, 0x44, 0x61, 0x79, 0x88, 0x01, 0x01, 0x12, 0x1f,
	0x0a, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x08,
	0x48, 0x0d, 0x52, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x88, 0x01, 0x01, 0x42,
	0x0a, 0x0a, 0x08, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x42, 0x05, 0x0a, 0x03, 0x5f,
	0x65, 0x71, 0x42, 0x05, 0x0a, 0x03, 0x5f, 0x6e, 0x65, 0x42, 0x05, 0x0a, 0x03, 0x5f, 0x6c, 0x74,
	0x42, 0x05, 0x0a, 0x03, 0x5f, 0x6c, 0x65, 0x42, 0x05, 0x0a, 0x03, 0x5f, 0x67, 0x74, 0x42, 0x05,
	0x0a, 0x03, 0x5f, 0x67, 0x65, 0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x6d, 0x61, 0x78, 0x5f, 0x63, 0x61,
	0x70, 0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x6d, 0x69, 0x6e, 0x5f, 0x63, 0x61, 0x70, 0x42, 0x0f, 0x0a,
	0x0d, 0x5f, 0x63, 0x73, 0x74, 0x5f, 0x64, 0x61, 0x74, 0x65, 0x74, 0x69, 0x6d, 0x65, 0x42, 0x0d,
	0x0a, 0x0b, 0x5f, 0x63, 0x73, 0x74, 0x5f, 0x6d, 0x69, 0x6e, 0x75, 0x74, 0x65, 0x42, 0x0c, 0x0a,
	0x0a, 0x5f, 0x63, 0x6e, 0x5f, 0x6d, 0x6f, 0x62, 0x69, 0x6c, 0x65, 0x42, 0x0a, 0x0a, 0x08, 0x5f,
	0x63, 0x73, 0x74, 0x5f, 0x64, 0x61, 0x79, 0x42, 0x0b, 0x0a, 0x09, 0x5f, 0x64, 0x75, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x22, 0x49, 0x0a, 0x0e, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x56, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x12, 0x26, 0x0a, 0x0c, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x0b,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x88, 0x01, 0x01, 0x42, 0x0f,
	0x0a, 0x0d, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x3a,
	0x53, 0x0a, 0x05, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x1d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64,
	0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xe8, 0xc6, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x19, 0x2e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x46, 0x69, 0x65, 0x6c,
	0x64, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x52, 0x05, 0x66, 0x69, 0x65, 0x6c,
	0x64, 0x88, 0x01, 0x01, 0x3a, 0x55, 0x0a, 0x05, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x12, 0x1f, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xe9,
	0xc6, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x6f, 0x72, 0x2e, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f,
	0x72, 0x52, 0x05, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x88, 0x01, 0x01, 0x42, 0x4c, 0x5a, 0x4a, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x6c, 0x75, 0x65, 0x6b, 0x61,
	0x6b, 0x69, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x76, 0x76, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2d, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f,
	0x72, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_validator_options_proto_rawDescOnce sync.Once
	file_validator_options_proto_rawDescData = file_validator_options_proto_rawDesc
)

func file_validator_options_proto_rawDescGZIP() []byte {
	file_validator_options_proto_rawDescOnce.Do(func() {
		file_validator_options_proto_rawDescData = protoimpl.X.CompressGZIP(file_validator_options_proto_rawDescData)
	})
	return file_validator_options_proto_rawDescData
}

var file_validator_options_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_validator_options_proto_goTypes = []interface{}{
	(*FieldValidator)(nil),              // 0: validator.FieldValidator
	(*MediaValidator)(nil),              // 1: validator.MediaValidator
	(*descriptorpb.FieldOptions)(nil),   // 2: google.protobuf.FieldOptions
	(*descriptorpb.MessageOptions)(nil), // 3: google.protobuf.MessageOptions
}
var file_validator_options_proto_depIdxs = []int32{
	2, // 0: validator.field:extendee -> google.protobuf.FieldOptions
	3, // 1: validator.media:extendee -> google.protobuf.MessageOptions
	0, // 2: validator.field:type_name -> validator.FieldValidator
	1, // 3: validator.media:type_name -> validator.MediaValidator
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	2, // [2:4] is the sub-list for extension type_name
	0, // [0:2] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_validator_options_proto_init() }
func file_validator_options_proto_init() {
	if File_validator_options_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_validator_options_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FieldValidator); i {
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
		file_validator_options_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MediaValidator); i {
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
	file_validator_options_proto_msgTypes[0].OneofWrappers = []interface{}{}
	file_validator_options_proto_msgTypes[1].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_validator_options_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 2,
			NumServices:   0,
		},
		GoTypes:           file_validator_options_proto_goTypes,
		DependencyIndexes: file_validator_options_proto_depIdxs,
		MessageInfos:      file_validator_options_proto_msgTypes,
		ExtensionInfos:    file_validator_options_proto_extTypes,
	}.Build()
	File_validator_options_proto = out.File
	file_validator_options_proto_rawDesc = nil
	file_validator_options_proto_goTypes = nil
	file_validator_options_proto_depIdxs = nil
}
