package sequential

import (
	"os"
	"testing"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/zaplog"

	"go.uber.org/zap"
)

var logger, _ = zaplog.NewJSONLogger()

func TestMain(m *testing.M) {
	defer logger.Sync()

	m.Run()
}

func TestXX(t *testing.T) {
	sequential := New("/tmp/sequential", logger)
	defer sequential.Close()
	sequential.string()

	if false {
		t.Log(sequential.Write([]byte("hello world 001")))
		sequential.string()
		t.Log(sequential.Write([]byte("hello world 002")))
		sequential.string()
		t.Log(sequential.Write([]byte("hello world 003")))
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

			t.Log(sequential.Write(entry))
			sequential.string()
		}
	}

	if false {
		for k := uint64(1); k <= 26; k++ {
			raw, err := sequential.Get(k)
			if err != nil {
				t.Error(err)
				continue
			}

			t.Log(k, len(raw), string(raw[:10]))
		}
	}

	if false {
		path := "/tmp/sequential/1.mox"
		file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			logger.Fatal("", zap.Error(errors.Wrapf(err, "open file %s err", path)))
		}
		defer file.Close()

		file.WriteAt([]byte{0, 1, 2, 3}, dataOffset+(128<<10)*4)
	}

	if false {
		err := Info("/tmp/sequential")
		if err != nil {
			t.Fatal(err)
		}
	}
}
