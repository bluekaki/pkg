package marshaler

import (
	"io"

	"github.com/bluekaki/pkg/errors"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func NewWildcardMarshaler() runtime.Marshaler {
	return new(wildcard)
}

var _ runtime.Marshaler = (*wildcard)(nil)

type wildcard struct {
	runtime.Marshaler
}

func (w *wildcard) ContentType(_ interface{}) string {
	return "*"
}

func (w *wildcard) Unmarshal(data []byte, value interface{}) error {

	return errors.New("unable to unmarshal non wildcard field")
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

func (w *wildcard) Marshal(v interface{}) ([]byte, error) {

	return nil, errors.New("unable to marshal non wildcard field")
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
