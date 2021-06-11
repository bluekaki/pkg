package interceptor

import (
	"net/url"
)

func QueryUnescape(uri string) string {
	decodedUri, err := url.QueryUnescape(uri)
	if err != nil {
		return uri
	}

	return decodedUri
}
