package timeutil

import (
	"time"
	_ "time/tzdata"
)

var (
	cst *time.Location
)

func init() {
	// IANA Asia/Shanghai
	var err error
	if cst, err = time.LoadLocation("Asia/Shanghai"); err != nil {
		panic(err)
	}
}

// CST China Standard Time
func CST() *time.Location {
	return cst
}

// NowInCST return a time now in cst
func NowInCST() *time.Time {
	ts := time.Now().In(cst)
	return &ts
}
