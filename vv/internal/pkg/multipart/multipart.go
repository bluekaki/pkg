package multipart

import (
	"io"
	"mime"
	"net/http"
	"net/textproto"

	"github.com/bluekaki/pkg/errors"
)

func Parse(req *http.Request) ([]byte, bool, error) {
	contentType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		return nil, false, errors.Wrap(err, "read Content-Type from header err")
	}

	switch textproto.CanonicalMIMEHeaderKey(contentType) {
	case "multipart/form-data":
		body, err := parseFormData(req)
		return body, true, err
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, false, errors.Wrap(err, "read request body err")
	}

	return body, false, err
}
