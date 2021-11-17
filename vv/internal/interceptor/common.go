package interceptor

import (
	"google.golang.org/grpc/metadata"
)

const (
	// JournalID a random id used by log journal
	JournalID = "journal-id"
	// Authorization used by auth, both gateway and grpc
	Authorization = "authorization"
	// AuthorizationProxy used by signature, both gateway and grpc
	AuthorizationProxy = "authorization-proxy"
	// Date GMT format
	Date = "date"
	// Method http.XXXMethod
	Method = "method"
	// URI url encoded
	URI = "uri"
	// Body string body
	Body = "body"
	// XForwardedFor forwarded for
	XForwardedFor = "x-forwarded-for"
	// XForwardedHost forwarded host
	XForwardedHost = "x-forwarded-host"
	// OctetStream binary files
	OctetStream = "octet-stream"
)

var toLoggedMetadata = map[string]bool{
	Authorization:      true,
	AuthorizationProxy: true,
	Date:               true,
	Method:             true,
	URI:                true,
	Body:               true,
	XForwardedFor:      true,
	XForwardedHost:     true,
	OctetStream:        true,
}

var gwHeader = struct {
	key   string
	value string
}{
	key:   "grpc-gateway",
	value: "bluekaki/pkg/vv",
}

func forwardedByGrpcGateway(meta metadata.MD) bool {
	values := meta.Get(gwHeader.key)
	if len(values) == 0 {
		return false
	}

	return values[0] == gwHeader.value
}
