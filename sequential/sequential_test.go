package sequential

import (
	"testing"

	"github.com/bluekaki/pkg/zaplog"
)

var logger, _ = zaplog.NewJSONLogger()

func TestMain(m *testing.M) {
	defer logger.Sync()

	m.Run()
}

func TestXX(t *testing.T) {
	sequential := New("/opt/tmp/sequential", logger)
	sequential.string()

	if false {
		sequential.Write([]byte("hello world 001"))
		sequential.string()
		sequential.Write([]byte("hello world 002"))
		sequential.string()
		sequential.Write([]byte("hello world 003"))
		sequential.string()
	}

	if false {
		entry := make([]byte, 128<<10)
		filling := func(char byte) {
			for i := range entry {
				entry[i] = char
			}
		}

		for k := byte('A'); k <= byte('Z'); k++ {
			filling(k)

			sequential.Write(entry)
			sequential.string()
		}
	}
}
