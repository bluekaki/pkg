package multipart

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"io"
	"net/http"

	"github.com/bluekaki/pkg/errors"
)

const (
	offsetSize   = 4
	boundarySize = 30
)

func parseFormData(req *http.Request) ([]byte, error) {
	if err := req.ParseMultipartForm(1 << 20); err != nil {
		return nil, errors.WithStack(err)
	}

	if len(req.MultipartForm.File) == 0 {
		return nil, errors.New("no file found in form-data")
	}

	var raw [][]byte
	for fieldName, files := range req.MultipartForm.File {
		if len(files) == 0 {
			return nil, errors.Errorf("no file found in form-data field of [%s]", fieldName)
		}

		for i := range files {
			err := func() error {
				file, err := files[i].Open()
				if err != nil {
					return errors.Wrapf(err, "open %s of field [%s] form form-data err", files[i].Filename, fieldName)
				}
				defer file.Close()

				body, err := io.ReadAll(file)
				if err != nil {
					return errors.Wrapf(err, "read %s of field [%s] form form-data err", files[i].Filename, fieldName)
				}

				raw = append(raw, body)
				return nil
			}()
			if err != nil {
				return nil, err
			}
		}
	}

	boundary := make([]byte, boundarySize)
	io.ReadFull(rand.Reader, boundary)
	payload := bytes.Join(raw, boundary)

	if len(raw) > 1 {
		offset := make([]byte, offsetSize)
		binary.BigEndian.PutUint32(offset, uint32(len(raw[0])))
		payload = append(offset, payload...)
	}

	return payload, nil
}

func ParseFormData(payload []byte) [][]byte {
	if len(payload) <= offsetSize+boundarySize {
		return [][]byte{payload}
	}

	offset := int(binary.BigEndian.Uint32(payload))
	if len(payload) <= offsetSize+offset+boundarySize {
		return [][]byte{payload}
	}

	return bytes.Split(payload[offsetSize:], payload[offsetSize+offset:offsetSize+offset+boundarySize])
}
