package sequential

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"
	"sync"
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

func TestManual(t *testing.T) {
	const baseDir = "/tmp"

	sequential := New(baseDir, logger)
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

			sequential.Write(entry)
			// sequential.string()
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
		path := baseDir + "/1.mox"
		file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			logger.Fatal("", zap.Error(errors.Wrapf(err, "open file %s err", path)))
		}
		defer file.Close()

		file.WriteAt([]byte{0, 1, 2, 3}, dataOffset+(128<<10)*4)
	}

	if false {
		err := Info(baseDir)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func BenchmarkSequential(b *testing.B) {
	b.SetParallelism(1)

	const total = 50 * 10000
	raw := make([]byte, 128<<10)

	const baseDir = "/data/sequential"

	digestSlice := make([][]byte, total+1)
	for k := uint64(1); k <= total; k++ {
		copy(raw, []byte(fmt.Sprintf("%08d", k)))

		digest := sha1.Sum(raw)
		digestSlice[k] = digest[:]

		if k%1000 == 0 {
			logger.Info("init digest", zap.Uint64("k", k))
		}
	}
	logger.Info("init digest over")

	sequential := New(baseDir, logger)
	defer sequential.Close()

	for k := uint64(1); k <= total; k++ {
		copy(raw, []byte(fmt.Sprintf("%08d", k)))

		offset, err := sequential.Write(raw)
		if err != nil {
			logger.Fatal(fmt.Sprintf("write into sequential %d err", k), zap.Error(err))
		}

		if offset != k {
			logger.Fatal("write into sequential return un-expected offset", zap.Uint64("k", k), zap.Uint64("offset", offset))
		}

		if k%1000 == 0 {
			logger.Info("write into sequential", zap.Uint64("k", k))
		}
	}
	logger.Info("write into sequential over")

	workers := 10
	ch := make(chan uint64, workers)
	wg := new(sync.WaitGroup)

	wg.Add(workers)
	for k := 0; k < workers; k++ {
		go func() {
			defer wg.Done()

			for offset := range ch {
				raw, err := sequential.Get(offset)
				if err != nil {
					logger.Fatal(fmt.Sprintf("get offset %d from sequential err", offset), zap.Error(err))
				}

				digest := sha1.Sum(raw)
				if !bytes.Equal(digest[:], digestSlice[offset]) {
					logger.Fatal("digest not match", zap.Uint64("offset", offset))
				}

				if offset%1000 == 0 {
					logger.Info("get from sequential", zap.Uint64("offset", offset))
				}
			}
		}()
	}

	for k := uint64(1); k <= total; k++ {
		ch <- k
	}
	close(ch)

	wg.Wait()
	logger.Info("over")

	Info(baseDir)
}
