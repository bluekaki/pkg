package pbutil

// Wire2String parse protobuf wire raw and encode to string
// https://developers.google.com/protocol-buffers/docs/encoding
// https://developers.google.com/protocol-buffers/docs/proto3#scalar
func Wire2String(raw []byte) (string, error) {
	if len(raw) == 0 {
		return "", nil
	}

	return "", nil
}
