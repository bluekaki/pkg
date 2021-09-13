package bpt

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc64"
	"os"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/stringutil"

	"go.uber.org/zap"
)

type Json2Value func([]byte) Value

type Value interface {
	String() string
	Compare(Value) stringutil.Diff
	ToJSON() []byte
}

type node struct {
	index    uint64
	values   []Value
	children []*node
}

type nodeSnapshot struct {
	Values   []json.RawMessage `json:"values"`
	Children []uint64          `json:"children"`
}

func (n *node) full(N int) bool {
	return len(n.values) == N
}

func (n *node) overHalf(HT int) bool {
	return len(n.values) >= HT
}

func (n *node) leaf() bool {
	return len(n.children) == 0
}

func (n *node) delete(baseDir string, logger *zap.Logger) {
	path := fmt.Sprintf("%s/%d%s", baseDir, n.index, fileExt)
	if err := os.Remove(path); err != nil {
		logger.Fatal("", zap.Error(errors.Wrapf(err, "delete file %s err", path)))
	}
}

func (n *node) takeSnapshots(baseDir string, logger *zap.Logger) {
	snapshot := &nodeSnapshot{
		Values:   make([]json.RawMessage, len(n.values)),
		Children: make([]uint64, len(n.children)),
	}

	for index, value := range n.values {
		snapshot.Values[index] = json.RawMessage(value.ToJSON())
	}

	for index, child := range n.children {
		snapshot.Children[index] = child.index
	}

	raw, err := json.Marshal(snapshot)
	if err != nil {
		logger.Fatal("", zap.Error(errors.Wrapf(err, "marshal node %d to json err", n.index)))
	}

	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(raw)))

	crc := make([]byte, 8)
	binary.BigEndian.PutUint64(crc, crc64.Checksum(raw, crc64.MakeTable(crc64.ECMA)))

	path := fmt.Sprintf("%s/%d%s", baseDir, n.index, fileExt)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		logger.Fatal("", zap.Error(errors.Wrapf(err, "open file %s err", path)))
	}

	if _, err = file.Write(length); err != nil {
		logger.Fatal("", zap.Error(errors.Wrapf(err, "write lenght into file %s err", path)))
	}
	if _, err = file.Write(raw); err != nil {
		logger.Fatal("", zap.Error(errors.Wrapf(err, "write raw into file %s err", path)))
	}
	if _, err = file.Write(crc); err != nil {
		logger.Fatal("", zap.Error(errors.Wrapf(err, "write crc into file %s err", path)))
	}

	if err = file.Sync(); err != nil {
		logger.Fatal("", zap.Error(errors.Wrapf(err, "sync file %s err", path)))
	}
	if err = file.Close(); err != nil {
		logger.Fatal("", zap.Error(errors.Wrapf(err, "close file %s err", path)))
	}
}

func (n *node) shouldLoadSnapshots() bool {
	return len(n.values) == 0

}

func loadSnapshots(index uint64, baseDir string, logger *zap.Logger, json2Value Json2Value) *node {
	path := fmt.Sprintf("%s/%d%s", baseDir, index, fileExt)
	raw, err := os.ReadFile(path)
	if err != nil {
		logger.Fatal("", zap.Error(errors.Wrapf(err, "read file %s err", path)))
	}

	snapshot := new(nodeSnapshot)
	if err = json.Unmarshal(raw, snapshot); err != nil {
		logger.Fatal("", zap.Error(errors.Wrapf(err, "unmarshal file %s err", path)))
	}

	n := &node{
		index:    index,
		values:   make([]Value, len(snapshot.Values)),
		children: make([]*node, len(snapshot.Children)),
	}

	for index, value := range snapshot.Values {
		n.values[index] = json2Value([]byte(value))
	}

	for index, child := range snapshot.Children {
		n.children[index] = &node{index: child}
	}

	return n
}
