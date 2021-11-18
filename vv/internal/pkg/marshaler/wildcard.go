package marshaler

import (
	"io"

	"github.com/bluekaki/pkg/errors"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
)

func NewWildcardMarshaler(logger *zap.Logger) runtime.Marshaler {
	if logger == nil {
		panic("logger required")
	}

	return &wildcard{
		logger: logger,
	}
}

type wildcard struct {
	runtime.Marshaler
	logger *zap.Logger
}

func (w *wildcard) NewDecoder(reader io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(func(value interface{}) error {
		buffer, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		return w.Unmarshal(buffer, value)
	})
}

func (w *wildcard) NewEncoder(writer io.Writer) runtime.Encoder {
	return runtime.EncoderFunc(func(value interface{}) error {
		buffer, err := w.Marshal(value)
		if err != nil {
			return err
		}
		_, err = writer.Write(buffer)
		if err != nil {
			return err
		}

		return nil
	})
}

func (w *wildcard) Unmarshal(data []byte, value interface{}) error {
	switch value.(type) {
	case *[]byte:
		message := value.(*[]byte)
		*message = make([]byte, len(data))
		copy(*message, data)

	default:
		err := errors.Errorf("wildcard unable to unmarshal type of %#v", value)
		w.logger.Error("wildcard unmarshal err", zap.Error(err))
		return err
	}

	return nil
}

type media interface {
	ContentType() string
	Payload() []byte
}

func (w *wildcard) ContentType(value interface{}) string {
	switch value.(type) {
	case media:
		return value.(media).ContentType()
	}

	return runtime.MIMEWildcard
}

func (w *wildcard) Marshal(value interface{}) ([]byte, error) {
	switch value.(type) {
	case media:
		return value.(media).Payload(), nil

	default:
		return jsonPbMarshaler.Marshal(value)
	}
}
