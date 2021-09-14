package bpt

import (
	"bytes"
	"container/list"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/bluekaki/pkg/errors"

	"go.uber.org/zap"
)

const (
	fileExt = ".bpt"

	rootIndex = 0
)

type bpTree struct {
	sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	logger *zap.Logger
	size   uint64
	root   *node

	json2Value    Json2Value
	loadSnapshots loadSnapshots

	meta struct {
		baseDir string
		index   uint64
		N       int // the max values in one node
		Mid     int // N / 2
		HT      int // (N + 1) / 2  (half order)the half children in one node
	}
}

func New(orderT uint16, baseDir string, logger *zap.Logger, json2Value Json2Value) *bpTree {
	if logger == nil {
		panic("logger required")
	}
	if json2Value == nil {
		panic("json2Value required")
	}

	if orderT%2 != 0 {
		logger.Fatal("", zap.Error(errors.New("t must be even number")))
	}

	if orderT < 4 { // t ≥4
		logger.Fatal("", zap.Error(errors.New("t must be ≥4")))
	}

	info, err := os.Stat(baseDir)
	if err != nil {
		logger.Fatal("", zap.Error(errors.Wrapf(err, "read dir %s stat err", baseDir)))
	}
	if !info.IsDir() {
		logger.Fatal("", zap.Error(errors.Errorf("%s should be directory", baseDir)))
	}

	baseDir = strings.TrimRight(strings.ReplaceAll(baseDir, "\\", "/"), "/")

	ctx, cancel := context.WithCancel(context.Background())
	bpt := &bpTree{
		ctx:           ctx,
		cancel:        cancel,
		logger:        logger,
		json2Value:    json2Value,
		loadSnapshots: loadSnapshotsBuilder(baseDir, logger, json2Value),
	}
	bpt.meta.baseDir = baseDir
	bpt.meta.N = int(orderT - 1)
	bpt.meta.Mid = int(orderT-1) / 2
	bpt.meta.HT = int(orderT) / 2

	bpt.init()
	return bpt
}

func (t *bpTree) Close() {
	select {
	case <-t.ctx.Done():
	default:
		t.cancel()

		// TODO ...
	}
}

func (t *bpTree) init() {
	err := filepath.Walk(t.meta.baseDir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() || filepath.Ext(info.Name()) != fileExt {
			return nil
		}

		fileIndex, err := strconv.ParseUint(info.Name()[:len(info.Name())-len(fileExt)], 10, 64)
		if err != nil {
			t.logger.Fatal("", zap.Error(errors.Wrapf(err, "parse file name of %s err", path)))
		}

		if fileIndex > t.meta.index {
			t.meta.index = fileIndex
		}

		return nil
	})
	if err != nil {
		t.logger.Fatal("", zap.Error(errors.Wrapf(err, "walk directory %s err", t.meta.baseDir)))
	}
}

func (t *bpTree) nextIndex() uint64 {
	t.meta.index++
	return t.meta.index
}

func (t *bpTree) String() string {
	t.RLock()
	defer t.RUnlock()

	stack := list.New()
	if fileExists(rootIndex, t.meta.baseDir) {
		output(stack, t.loadSnapshots(rootIndex), 0, t.loadSnapshots)
	}

	buf := bytes.NewBufferString(fmt.Sprintf("BTree %d\n", t.size))
	for stack.Len() > 0 {
		element := stack.Back()
		stack.Remove(element)

		buf.WriteString(element.Value.(string))
	}

	return buf.String()
}

func output(stack *list.List, node *node, level int, load loadSnapshots) {
	for e := 0; e < len(node.values)+1; e++ {
		if e < len(node.children) {
			output(stack, load(node.children[e].index), level+1, load)
		}

		if e < len(node.values) {
			stack.PushBack(fmt.Sprintf("%s%s(%d)\n", strings.Repeat("    ", level), node.values[e].String(), node.index))
		}
	}
}

func (t *bpTree) Empty() bool {
	t.RLock()
	defer t.RUnlock()

	return t.size == 0
}

func (t *bpTree) Size() uint64 {
	t.RLock()
	defer t.RUnlock()

	return t.size
}

func (t *bpTree) Asc() (values []Value) {
	t.RLock()
	defer t.RUnlock()

	if !fileExists(rootIndex, t.meta.baseDir) {
		return
	}

	type item struct {
		*node
		cur int
	}

	stack := list.New()
	stack.PushBack(&item{node: t.loadSnapshots(rootIndex)})

	for stack.Len() > 0 {
		element := stack.Back()
		stack.Remove(element)

		node := element.Value.(*item)
		if node.node != t.root && len(node.values) < (t.meta.HT-1) {
			t.logger.Fatal("", zap.Error(errors.Errorf("illegal %d %s", node.node.index, t.String())))
		}

		if node.leaf() {
			values = append(values, node.values...)
			continue
		}

		if node.cur <= len(node.values) {
			child := &item{node: t.loadSnapshots(node.children[node.cur].index)}

			if node.cur > 0 {
				values = append(values, node.values[node.cur-1])
			}
			node.cur++
			stack.PushBack(node)
			stack.PushBack(child)
		}
	}
	return
}
