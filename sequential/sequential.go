package sequential

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/binary"
	stderr "errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bluekaki/pkg/errors"

	"go.uber.org/zap"
)

/*
|----------------------------------|
| meta:                            |
| cap     ts0        tsN           |
| xxxx~2021/01/01~2021/04/04       |
|                         26byte   |
|----------------------------------|
| index:                           |
| offset  length  digest  summary  |
| uint64  uint32  sha1             |
| 8byte   4byte   20byte           |
|                         32byte   |
|----------------------------------|
| file:                            |
| 26B(meta)+256K(index)+raw(data)  |
|           32B*8192    128K*8189  |
|----------------------------------|
*/

var (
	ErrOversize = fmt.Errorf("payload size exceeds the limit of %d", dataSize)
	ErrClosed   = stderr.New("sequential has closed")
	ErrNotfound = stderr.New("offset not found")
)

const (
	fileExt = ".mox"

	capOffset   = 0
	ts0Offset   = 5
	tsNOffset   = 16
	indexOffset = 26
	dataOffset  = indexOffset + indexSize

	metaSize = indexOffset

	indexLen  = 32
	indexSize = 256 << 10 // 256Kb

	fileSize = 1 << 30 // 1Gb
	dataSize = fileSize - dataOffset

	dateLayout = "2006/01/02"
)

// Sequential write files sequentially
type Sequential interface {
	Close()
	Write(raw []byte) (offset uint64, err error)
	Get(offset uint64) ([]byte, error)

	string()
}

type payload struct {
	offset uint64
	raw    []byte
	done   chan struct{}
}

var _ Sequential = (*sequential)(nil)

type sequential struct {
	ctx     context.Context
	cancel  context.CancelFunc
	baseDir string
	logger  *zap.Logger
	blocks  *blocks
	ch      chan *payload

	meta struct {
		minOffset   uint64
		maxOffset   uint64
		indexOffset int64
		dataOffset  int64

		fileIndex uint64
		file      *os.File
	}
}

// New write files sequentially
func New(baseDir string, logger *zap.Logger) Sequential {
	if logger == nil {
		panic("logger required")
	}

	info, err := os.Stat(baseDir)
	if err != nil {
		logger.Fatal("", zap.Error(errors.Wrapf(err, "read dir %s stat err", baseDir)))
	}

	if !info.IsDir() {
		logger.Fatal("", zap.Error(errors.New(baseDir+" should be directory")))
	}

	ctx, cancel := context.WithCancel(context.Background())
	sequential := &sequential{
		ctx:     ctx,
		cancel:  cancel,
		baseDir: strings.TrimRight(strings.ReplaceAll(baseDir, "\\", "/"), "/"),
		logger:  logger,
		blocks:  newBlocks(),
		ch:      make(chan *payload, 10),
	}

	sequential.validate()
	go sequential.consumer()

	return sequential
}

func (s *sequential) Close() {
	select {
	case <-s.ctx.Done():
	default:
		s.cancel()

		s.meta.file.Close()
		s.blocks.Close()
	}
}

func (s *sequential) string() {
	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	fmt.Println("fileIndex:", s.meta.fileIndex, "minOffset:", s.meta.minOffset, "maxOffset:", s.meta.maxOffset, "indexOffset:", s.meta.indexOffset, "dataOffset:", s.meta.dataOffset)
	s.blocks.String()
	fmt.Println("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
}

func (s *sequential) rdonly(path string) *os.File {
	file, err := os.Open(path)
	if err != nil {
		s.logger.Fatal("", zap.Error(errors.Wrapf(err, "read file %s err", path)))
	}

	return file
}

func (s *sequential) rdwr(path string) *os.File {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		s.logger.Fatal("", zap.Error(errors.Wrapf(err, "open file %s err", path)))
	}

	return file
}

func (s *sequential) validate() {
	err := filepath.Walk(s.baseDir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() || filepath.Ext(info.Name()) != fileExt {
			return nil
		}

		fileIndex, err := strconv.ParseUint(info.Name()[:len(info.Name())-len(fileExt)], 10, 64)
		if err != nil {
			return errors.Wrapf(err, "parse file name of %s err", path)
		}

		file := s.rdonly(path)

	redo:
		var metaRaw [metaSize]byte
		if _, err = file.ReadAt(metaRaw[:], capOffset); err != nil {
			s.logger.Fatal("", zap.Error(errors.Wrapf(err, "read meta of file %s err", path)))
		}

		var indexRaw [indexSize]byte
		if _, err = file.ReadAt(indexRaw[:], indexOffset); err != nil {
			s.logger.Fatal("", zap.Error(errors.Wrapf(err, "read index of file %s err", path)))
		}

		next := reversedIndex(indexRaw)
		loop := 0
		for {
			minOffset, maxOffset, digest0, index, err := next()
			if err != nil {
				break
			}

			idx := index.Last()
			raw := make([]byte, idx.Length())
			if _, err = file.ReadAt(raw, dataOffset+idx.DataOffset()); err != nil {
				s.logger.Fatal("", zap.Error(errors.Wrapf(err, "read offset %d in file %s err", idx.Offset(), path)))
			}

			digest1 := sha1.Sum(raw)
			if !bytes.Equal(digest1[:], digest0) {
				loop++
				continue
			}

			if loop > 0 {
				s.eraseInvalidIndex(minOffset, maxOffset, path)
				goto redo
			}

			s.blocks.AppendAndSort(&block{
				fileIndex: fileIndex,
				file:      file,
				minOffset: minOffset,
				maxOffset: maxOffset,
				index:     index,
			})
			break
		}

		return nil
	})
	if err != nil {
		s.logger.Fatal("", zap.Error(errors.Wrapf(err, "walk directory %s err", s.baseDir)))
	}

	last := s.blocks.Last()
	if last == nil {
		s.createFile()
		return
	}

	s.meta.minOffset = last.minOffset
	s.meta.maxOffset = last.maxOffset

	if last.maxOffset > 0 {
		s.meta.indexOffset = (int64(last.maxOffset-last.minOffset) + 1) * indexLen

		entry := last.index.Last()
		s.meta.dataOffset = entry.DataOffset() + int64(entry.Length())
	}

	s.meta.fileIndex = last.fileIndex
	s.meta.file = s.rdwr(last.file.Name())
}

func (s *sequential) eraseInvalidIndex(minOffset, maxOffset uint64, path string) {
	offset := int64((maxOffset - minOffset + 1) * indexLen)
	empty := make([]byte, indexSize-offset)

	file := s.rdwr(path)

	if _, err := file.WriteAt(encodeCapacity(minOffset, maxOffset), capOffset); err != nil {
		s.logger.Fatal("", zap.Error(errors.Wrapf(err, "write capacity into file %s err", path)))
	}

	if _, err := file.WriteAt([]byte(time.Now().Format(dateLayout)), tsNOffset); err != nil {
		s.logger.Fatal("", zap.Error(errors.Wrapf(err, "write tsN into file %s err", path)))
	}

	if _, err := file.WriteAt(empty, indexOffset+offset); err != nil {
		s.logger.Fatal("", zap.Error(errors.Wrapf(err, "write empty index into file %s err", path)))
	}

	if err := file.Sync(); err != nil {
		s.logger.Fatal("", zap.Error(errors.Wrapf(err, "sync file %s err", path)))
	}

	if err := file.Close(); err != nil {
		s.logger.Fatal("", zap.Error(errors.Wrapf(err, "close file %s err", path)))
	}
}

func (s *sequential) createEmptyFile(path string) (file *os.File, err error) {
	defer func() {
		if err != nil && file != nil {
			file.Close()
		}
	}()

	file = s.rdwr(path)

	delimiter := []byte{'~'}
	ts := []byte(time.Now().Format(dateLayout))

	if _, err = file.WriteAt(delimiter, ts0Offset-1); err != nil {
		return nil, errors.Wrapf(err, "write ts0 delimiter into file %s err", path)
	}

	if _, err = file.WriteAt(ts, ts0Offset); err != nil {
		return nil, errors.Wrapf(err, "write ts0 into file %s err", path)
	}

	if _, err = file.WriteAt(delimiter, tsNOffset-1); err != nil {
		return nil, errors.Wrapf(err, "write tsN delimiter into file %s err", path)
	}

	if _, err = file.WriteAt(ts, tsNOffset); err != nil {
		return nil, errors.Wrapf(err, "write tsN into file %s err", path)
	}

	if _, err = file.WriteAt([]byte{0}, fileSize-1); err != nil {
		return nil, errors.Wrapf(err, "write last zero byte into file %s err", path)
	}

	if err = file.Sync(); err != nil {
		return nil, errors.Wrapf(err, "sync file %s err", path)
	}

	return
}

func (s *sequential) createFile() {
	s.meta.minOffset = 0
	s.meta.maxOffset = 0
	s.meta.indexOffset = 0
	s.meta.dataOffset = 0
	s.meta.fileIndex++

	path := fmt.Sprintf("%s/%d%s", s.baseDir, s.meta.fileIndex, fileExt)
	file, err := s.createEmptyFile(path)
	if err != nil {
		s.logger.Fatal("", zap.Error(err))
	}
	s.meta.file = file

	s.blocks.Append(&block{
		fileIndex: s.meta.fileIndex,
		file:      s.rdonly(s.meta.file.Name()),
		index:     newIndex(),
	})
}

// Write if err occurs ErrOversize and ErrClosed will be returned
func (s *sequential) Write(raw []byte) (offset uint64, err error) {
	if len(raw) > dataSize {
		return 0, ErrOversize
	}

	payload := &payload{
		raw:  raw,
		done: make(chan struct{}),
	}
	s.ch <- payload

	select {
	case <-s.ctx.Done():
		return 0, ErrClosed

	case <-payload.done:
		return payload.offset, nil
	}
}

func (s *sequential) consumer() {
	defer func() {
		recover() // just ignore
	}()

	for {
		select {
		case <-s.ctx.Done():
			return

		case payload := <-s.ch:
			if (s.meta.indexOffset+indexLen) > indexSize || (s.meta.dataOffset+int64(len(payload.raw))) > dataSize {
				payload.offset, _ = s.rotateFile(payload.raw)
				close(payload.done)
				continue
			}

			if s.meta.minOffset == 0 {
				s.meta.minOffset++
			}
			s.meta.maxOffset++
			payload.offset = s.meta.maxOffset

			s.write(payload.raw, payload.offset)
			close(payload.done)
		}
	}
}

func (s *sequential) write(raw []byte, offset uint64) {
	index := make([]byte, indexLen)
	binary.BigEndian.PutUint64(index[:8], offset)
	binary.BigEndian.PutUint32(index[8:12], uint32(len(raw)))

	digest := sha1.Sum(raw)
	copy(index[12:], digest[:])

	if _, err := s.meta.file.WriteAt(encodeCapacity(s.meta.minOffset, s.meta.maxOffset), capOffset); err != nil {
		s.logger.Fatal("", zap.Error(errors.Wrapf(err, "write capacity into file %s err", s.meta.file.Name())))
	}

	if _, err := s.meta.file.WriteAt([]byte(time.Now().Format(dateLayout)), tsNOffset); err != nil {
		s.logger.Fatal("", zap.Error(errors.Wrapf(err, "write tsN into file %s err", s.meta.file.Name())))
	}

	if _, err := s.meta.file.WriteAt(index, indexOffset+s.meta.indexOffset); err != nil {
		s.logger.Fatal("", zap.Error(errors.Wrapf(err, "write index into file %s err", s.meta.file.Name())))
	}

	if _, err := s.meta.file.WriteAt(raw, dataOffset+s.meta.dataOffset); err != nil {
		s.logger.Fatal("", zap.Error(errors.Wrapf(err, "write data into file %s err", s.meta.file.Name())))
	}

	if err := s.meta.file.Sync(); err != nil {
		s.logger.Fatal("", zap.Error(errors.Wrapf(err, "sync file %s err", s.meta.file.Name())))
	}

	s.blocks.UpdateLast(newEntry(offset, s.meta.dataOffset, len(raw)), s.meta.minOffset, s.meta.maxOffset)

	s.meta.indexOffset += indexLen
	s.meta.dataOffset += int64(len(raw))
}

func (s *sequential) rotateFile(raw []byte) (offset uint64, err error) {
	offset = s.meta.maxOffset + 1

	if err := s.meta.file.Sync(); err != nil {
		s.logger.Fatal("", zap.Error(errors.Wrapf(err, "sync file %s err", s.meta.file.Name())))
	}

	if err := s.meta.file.Close(); err != nil { // close writer fd
		s.logger.Fatal("", zap.Error(errors.Wrapf(err, "close file %s err", s.meta.file.Name())))
	}

	s.createFile()
	s.meta.minOffset = offset
	s.meta.maxOffset = offset

	s.write(raw, offset)
	return
}

// Get if err occurs ErrNotfound and ErrClosed will be returned;
// notice: if read the underlayer file err, this truth err will be returned.
func (s *sequential) Get(offset uint64) ([]byte, error) {
	select {
	case <-s.ctx.Done():
		return nil, ErrClosed

	default:
		return s.blocks.Get(offset)
	}
}

// Info show offset and available capacity of each file(.mox) in this directory
func Info(baseDir string) error {
	info, err := os.Stat(baseDir)
	if err != nil {
		return errors.Wrapf(err, "read dir %s stat err", baseDir)
	}

	if !info.IsDir() {
		return errors.New(baseDir + " should be directory")
	}

	baseDir = strings.TrimRight(strings.ReplaceAll(baseDir, "\\", "/"), "/")

	var fileIndex []uint64
	err = filepath.Walk(baseDir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() || filepath.Ext(info.Name()) != fileExt {
			return nil
		}

		index, err := strconv.ParseUint(info.Name()[:len(info.Name())-len(fileExt)], 10, 64)
		if err != nil {
			return errors.Wrapf(err, "parse file name of %s err", path)
		}

		fileIndex = append(fileIndex, index)
		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "walk directory %s err", baseDir)
	}

	sort.Slice(fileIndex, func(i, j int) bool {
		return fileIndex[i] < fileIndex[j]
	})

	for _, index := range fileIndex {
		path := fmt.Sprintf("%s/%d%s", baseDir, index, fileExt)
		err = func() error {
			file, err := os.Open(path)
			if err != nil {
				return errors.Wrapf(err, "read file %s err", path)
			}
			defer file.Close()

			var ts0 [10]byte
			if _, err = file.ReadAt(ts0[:], ts0Offset); err != nil {
				return errors.Wrapf(err, "read ts0 of file %s err", path)
			}

			var tsN [10]byte
			if _, err = file.ReadAt(tsN[:], ts0Offset); err != nil {
				return errors.Wrapf(err, "read tsN of file %s err", path)
			}

			var indexRaw [indexSize]byte
			if _, err = file.ReadAt(indexRaw[:], indexOffset); err != nil {
				return errors.Wrapf(err, "read index of file %s err", path)
			}

			next := reversedIndex(indexRaw)
			minOffset, maxOffset, _, index, err := next()

			var capacity, leftIndex, leftData uint64
			if err == nil {
				capacity = maxOffset - minOffset + 1
				leftIndex = (indexSize / indexLen) - capacity
				leftData = dataSize - uint64(index.Last().DataOffset()) - uint64(index.Last().Length())
			}

			fmt.Println(
				path,
				"Capacity:", capacity,
				"MinOffset:", minOffset,
				"MaxOffset:", maxOffset,
				"LeftIndex:", leftIndex,
				"LeftData:", fmt.Sprintf("%.2fKb  %.2fMb", float64(leftData)/1024., float64(leftData)/1024./1024.),
			)

			return nil
		}()
		if err != nil {
			return err
		}
	}

	return nil
}
