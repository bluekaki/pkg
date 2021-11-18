package marshaler

import (
	"io"

	"github.com/bluekaki/pkg/errors"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func NewFromDataMarshaler() runtime.Marshaler {
	return &fromData{Marshaler: jsonPbMarshaler}
}

type fromData struct {
	runtime.Marshaler
}

func (f *fromData) ContentType(_ interface{}) string {
	return "multipart/form-data"
}

func (f *fromData) Unmarshal(data []byte, value interface{}) error {
	message, ok := value.(*[]byte)
	if !ok {
		return errors.New("unable to unmarshal non bytes field")
	}

	*message = make([]byte, len(data))
	copy(*message, data)

	return nil
}

func (f *fromData) NewDecoder(reader io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(func(value interface{}) error {
		buffer, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		return f.Unmarshal(buffer, value)
	})
}
