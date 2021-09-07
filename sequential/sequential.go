package sequential

import (
	// "bytes"
	"crypto/sha1"
	"encoding/binary"
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

const (
	fileExt = ".mox"

	capOffset   = 0
	ts0Offset   = 5
	tsNOffset   = 16
	indexOffset = 26
	dataOffset  = indexOffset + indexSize

	indexLen  = 32
	indexSize = 256 << 10 // 256Kb

	fileSize = 1 << 20 // 1 << 30 // 1Gb
	dataSize = fileSize - dataOffset

	dateLayout = "2006/01/02"
)

type block struct {
	fileIndex uint64
	file      *os.File
	minOffset uint64
	maxOffset uint64
	indexRaw  []byte
}

type sequential struct {
	baseDir string
	logger  *zap.Logger
	blocks  []*block

	meta struct {
		minOffset   uint64
		maxOffset   uint64
		indexOffset int64
		dataOffset  int64

		fileIndex uint64
		file      *os.File
	}
}

func New(baseDir string, logger *zap.Logger) *sequential {
	if logger == nil {
		panic("logger required")
	}

	info, err := os.Stat(baseDir)
	if err != nil {
		logger.Fatal(fmt.Sprintf("read dir %s stat err", baseDir), zap.Error(err))
	}

	if !info.IsDir() {
		logger.Fatal(baseDir + " should be directory")
	}

	sequential := &sequential{
		baseDir: strings.TrimRight(strings.ReplaceAll(baseDir, "\\", "/"), "/"),
		logger:  logger,
	}

	sequential.validate()
	return sequential
}

func (s *sequential) Close() {
	if s.meta.file != nil {
		s.meta.file.Close()
	}

	for _, block := range s.blocks {
		block.file.Close()
	}
}

func (s *sequential) string() {
	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	fmt.Println("fileIndex:", s.meta.fileIndex, "minOffset:", s.meta.minOffset, "maxOffset:", s.meta.maxOffset, "indexOffset:", s.meta.indexOffset, "dataOffset:", s.meta.dataOffset)
	for i, block := range s.blocks {
		fmt.Println(i, "fileIndex:", block.fileIndex, "minOffset:", block.minOffset, "maxOffset:", block.maxOffset)
	}
	fmt.Println("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
}

func (s *sequential) appendBlock(block *block) {
	s.blocks = append(s.blocks, block)

	sort.Slice(s.blocks, func(i, j int) bool {
		return s.blocks[i].fileIndex < s.blocks[j].fileIndex
	})
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

		file, err := os.Open(path)
		if err != nil {
			s.logger.Fatal(fmt.Sprintf("read file %s err", path), zap.Error(err))
		}

		indexRaw := make([]byte, indexSize)
		if _, err = file.ReadAt(indexRaw, indexOffset); err != nil {
			s.logger.Fatal(fmt.Sprintf("read index of file %s err", path), zap.Error(err))
		}

		_, minOffset, maxOffset, _ := latestBlock(indexRaw)
		s.appendBlock(&block{
			fileIndex: fileIndex,
			file:      file,
			minOffset: minOffset,
			maxOffset: maxOffset,
			indexRaw:  indexRaw,
		})
		return nil
	})
	if err != nil {
		s.logger.Fatal(fmt.Sprintf("walk directory %s err", s.baseDir), zap.Error(err))
	}

	if len(s.blocks) == 0 {
		s.createFile()
		return
	}

	latest := s.blocks[len(s.blocks)-1]
	s.blocks = s.blocks[:len(s.blocks)-1]

	if err := latest.file.Close(); err != nil { // close reader fd
		s.logger.Fatal(fmt.Sprintf("close file %s err", latest.file.Name()), zap.Error(err))
	}

	file, err := os.OpenFile(latest.file.Name(), os.O_RDWR, 0644)
	if err != nil {
		s.logger.Fatal(fmt.Sprintf("open file %s err", latest.file.Name()), zap.Error(err))
	}

	_, minOffset, maxOffset, dataOffset := latestBlock(latest.indexRaw)

	s.meta.minOffset = minOffset
	s.meta.maxOffset = maxOffset

	if maxOffset > 0 {
		s.meta.indexOffset = (int64(maxOffset-minOffset) + 1) * indexLen
		s.meta.dataOffset = dataOffset
	}

	s.meta.fileIndex = latest.fileIndex
	s.meta.file = file
}

func (s *sequential) createFile() {
	s.meta.minOffset = 0
	s.meta.maxOffset = 0
	s.meta.indexOffset = 0
	s.meta.dataOffset = 0
	s.meta.fileIndex++

	path := fmt.Sprintf("%s/%d%s", s.baseDir, s.meta.fileIndex, fileExt)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		s.logger.Fatal(fmt.Sprintf("create file %s err", path), zap.Error(err))
	}
	s.meta.file = file

	buf := make([]byte, 4<<10)
	loop := fileSize / (4 << 10)
	for k := 0; k < loop; k++ {
		if _, err = s.meta.file.Write(buf); err != nil {
			s.logger.Fatal(fmt.Sprintf("write zero into file %s err", path), zap.Error(err))
		}
	}

	delimiter := []byte{'~'}
	ts := []byte(time.Now().Format(dateLayout))

	s.meta.file.WriteAt(delimiter, ts0Offset-1)
	s.meta.file.WriteAt(ts, ts0Offset)
	s.meta.file.WriteAt(delimiter, tsNOffset-1)
	s.meta.file.WriteAt(ts, tsNOffset)

	if err = s.meta.file.Sync(); err != nil {
		s.logger.Fatal(fmt.Sprintf("sync file %s err", path), zap.Error(err))
	}
}

func (s *sequential) Write(raw []byte) (offset uint64, err error) {
	if len(raw) > dataSize {
		return 0, errors.Errorf("payload size exceeds the limit of %d", dataSize)
	}

	if (s.meta.indexOffset+indexLen) > indexSize || (s.meta.dataOffset+int64(len(raw))) > dataSize {
		return s.rotateFile(raw)
	}

	if s.meta.minOffset == 0 {
		s.meta.minOffset++
	}
	s.meta.maxOffset++
	offset = s.meta.maxOffset

	s.write(raw, offset)
	return
}

func (s *sequential) write(raw []byte, offset uint64) {
	index := make([]byte, 32)
	binary.BigEndian.PutUint64(index[:8], offset)
	binary.BigEndian.PutUint32(index[8:12], uint32(len(raw)))

	digest := sha1.Sum(raw)
	copy(index[12:], digest[:])

	if _, err := s.meta.file.WriteAt([]byte(strconv.FormatUint(s.meta.maxOffset-s.meta.minOffset+1, 10)), capOffset); err != nil {
		s.logger.Fatal(fmt.Sprintf("write capacity into file %s err", s.meta.file.Name()), zap.Error(err))
	}

	if _, err := s.meta.file.WriteAt([]byte(time.Now().Format(dateLayout)), tsNOffset); err != nil {
		s.logger.Fatal(fmt.Sprintf("write tsN into file %s err", s.meta.file.Name()), zap.Error(err))
	}

	if _, err := s.meta.file.WriteAt(index, indexOffset+s.meta.indexOffset); err != nil {
		s.logger.Fatal(fmt.Sprintf("write index into file %s err", s.meta.file.Name()), zap.Error(err))
	}

	if _, err := s.meta.file.WriteAt(raw, dataOffset+s.meta.dataOffset); err != nil {
		s.logger.Fatal(fmt.Sprintf("write data into file %s err", s.meta.file.Name()), zap.Error(err))
	}

	if err := s.meta.file.Sync(); err != nil {
		s.logger.Fatal(fmt.Sprintf("sync file %s err", s.meta.file.Name()), zap.Error(err))
	}

	s.meta.indexOffset += indexLen
	s.meta.dataOffset += int64(len(raw))
}

func (s *sequential) rotateFile(raw []byte) (offset uint64, err error) {
	offset = s.meta.maxOffset + 1

	if err := s.meta.file.Sync(); err != nil {
		s.logger.Fatal(fmt.Sprintf("sync file %s err", s.meta.file.Name()), zap.Error(err))
	}

	if err := s.meta.file.Close(); err != nil { // close writer fd
		s.logger.Fatal(fmt.Sprintf("close file %s err", s.meta.file.Name()), zap.Error(err))
	}

	file, err := os.Open(s.meta.file.Name())
	if err != nil {
		s.logger.Fatal(fmt.Sprintf("read file %s err", s.meta.file.Name()), zap.Error(err))
	}

	indexRaw := make([]byte, indexSize)
	if _, err = file.ReadAt(indexRaw, indexOffset); err != nil {
		s.logger.Fatal(fmt.Sprintf("read index of file %s err", s.meta.file.Name()), zap.Error(err))
	}

	s.appendBlock(&block{
		fileIndex: s.meta.fileIndex,
		file:      file,
		minOffset: s.meta.minOffset,
		maxOffset: s.meta.maxOffset,
		indexRaw:  indexRaw,
	})

	s.createFile()
	s.meta.minOffset = offset
	s.meta.maxOffset = offset

	s.write(raw, offset)
	return
}
