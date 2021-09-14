package bpt

import (
	"fmt"
	"hash/crc64"
	"os"
)

func caclCrc(raw []byte) uint64 {
	return crc64.Checksum(raw, crc64.MakeTable(crc64.ECMA))
}

func fileExists(index uint64, baseDir string) bool {
	_, err := os.Stat(fmt.Sprintf("%s/%d%s", baseDir, index, fileExt))
	return err == nil
}
