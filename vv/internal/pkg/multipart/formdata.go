package multipart

import (
	"io"
	"net/http"

	"github.com/bluekaki/pkg/errors"
)

func parseFormData(req *http.Request) ([]byte, error) {
	if err := req.ParseMultipartForm(1 << 20); err != nil {
		return nil, errors.WithStack(err)
	}

	if len(req.MultipartForm.File) == 0 {
		return nil, errors.New("no file found in form-data")
	}

	for _, files := range req.MultipartForm.File {
		if len(files) == 0 {
			return nil, errors.New("no file found in form-data")
		}

		file, err := files[0].Open()
		if err != nil {
			return nil, errors.Wrapf(err, "open %s form form-data err", files[0].Filename)
		}
		defer file.Close()

		body, err := io.ReadAll(file)
		if err != nil {
			return nil, errors.Wrapf(err, "read %s form form-data err", files[0].Filename)
		}

		return body, nil
	}

	return nil, nil
}
