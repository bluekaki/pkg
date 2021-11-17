package gateway

import (
	"io"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/vv/internal/pkg/multipart"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type formData struct {
	runtime.JSONPb
}

func (f *formData) ContentType(_ interface{}) string {
	return multipart.ContentTypeFormData
}

func (f *formData) Unmarshal(data []byte, value interface{}) error {
	message, ok := value.(*[]byte)
	if !ok {
		return errors.New("unable to unmarshal non bytes field")
	}

	*message = make([]byte, len(data))
	copy(*message, data)

	return nil
}

func (f *formData) NewDecoder(reader io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(func(value interface{}) error {
		buffer, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		return f.Unmarshal(buffer, value)
	})
}
