package server

import (
	"github.com/bluekaki/pkg/vv/internal/interceptor"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// ParseFileDescriptor parse file descriptor
func ParseFileDescriptor(descriptor protoreflect.FileDescriptor) {
	if descriptor == nil {
		panic("file descriptor required")
	}

	interceptor.FileDescriptor.VerifyValidator(descriptor)
	interceptor.FileDescriptor.ParseMethod(descriptor, true)
}
