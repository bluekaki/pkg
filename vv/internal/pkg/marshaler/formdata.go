package marshaler

import (
	"io"

	"github.com/bluekaki/pkg/errors"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
)

func NewFormDataMarshaler(logger *zap.Logger) runtime.Marshaler {
	if logger == nil {
		panic("logger required")
	}

	return &formData{
		Marshaler: jsonPbMarshaler,
		logger:    logger,
	}
}

type formData struct {
	runtime.Marshaler
	logger *zap.Logger
}

func (f *formData) ContentType(_ interface{}) string {
	return "multipart/form-data"
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
