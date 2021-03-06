package bpt

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/stringutil"

	"go.uber.org/zap"
)

type Json2Value func([]byte) Value
type loadSnapshots func(index uint64) *node
type deleteNode func(*node)
type nodeTakeSnapshot func(*node)

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

func deleteNodeBuilder(baseDir string, logger *zap.Logger) deleteNode {
	return func(n *node) {
		path := fmt.Sprintf("%s/%d%s", baseDir, n.index, fileExt)
		if err := os.Remove(path); err != nil {
			logger.Fatal("", zap.Error(errors.Wrapf(err, "delete file %s err", path)))
		}
	}
}

func nodeTakeSnapshotBuilder(baseDir string, logger *zap.Logger) nodeTakeSnapshot {
	return func(n *node) {
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
		binary.BigEndian.PutUint64(crc, caclCrc(raw))

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
}

func (n *node) shouldLoadSnapshots() bool {
	return len(n.values) == 0
}

func loadSnapshotsBuilder(baseDir string, logger *zap.Logger, json2Value Json2Value) loadSnapshots {
	return func(index uint64) *node {
		path := fmt.Sprintf("%s/%d%s", baseDir, index, fileExt)
		file, err := os.Open(path)
		if err != nil {
			logger.Fatal("", zap.Error(errors.Wrapf(err, "open file %s err", path)))
		}
		defer file.Close()

		length := make([]byte, 4)
		if _, err = file.Read(length); err != nil {
			logger.Fatal("", zap.Error(errors.Wrapf(err, "read length from file %s err", path)))
		}

		raw := make([]byte, binary.BigEndian.Uint32(length))
		if _, err = file.Read(raw); err != nil {
			logger.Fatal("", zap.Error(errors.Wrapf(err, "read raw from file %s err", path)))
		}

		crc := make([]byte, 8)
		if _, err = file.Read(crc); err != nil {
			logger.Fatal("", zap.Error(errors.Wrapf(err, "read crc from file %s err", path)))
		}

		if binary.BigEndian.Uint64(crc) != caclCrc(raw) {
			logger.Fatal("", zap.Error(errors.Errorf("crc not match in file %s", path)))
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
}
