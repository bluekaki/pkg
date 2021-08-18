package interceptor

import (
	"net/url"
	"os"

	"github.com/bluekaki/pkg/pbutil"

	"github.com/golang/protobuf/proto"
)

func QueryUnescape(uri string) string {
	decodedUri, err := url.QueryUnescape(uri)
	if err != nil {
		return uri
	}

	return decodedUri
}

func marshalJournal(journal proto.Message) interface{} {
	raw, _ := pbutil.ProtoMessage2JSON(journal)

	if os.Getenv("MarshalJournal") == "true" {
		return string(raw)
	}
	return raw
}
