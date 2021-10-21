package interceptor

import (
	"net/url"
	"os"

	"github.com/bluekaki/pkg/pbutil"

	"github.com/golang/protobuf/proto"
)

func queryUnescape(uri string) string {
	decodedURI, err := url.QueryUnescape(uri)
	if err != nil {
		return uri
	}

	return decodedURI
}

func marshalJournal(journal proto.Message) interface{} {
	raw, _ := pbutil.ProtoMessage2JSON(journal)

	if os.Getenv("MarshalJournal") == "true" {
		return string(raw)
	}
	return raw
}
