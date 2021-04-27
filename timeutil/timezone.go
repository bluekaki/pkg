package timeutil

import (
	"time"
	_ "time/tzdata"
)

var (
	cst *time.Location
	jst *time.Location
)

func init() {
	// IANA Asia/Shanghai
	var err error
	if cst, err = time.LoadLocation("Asia/Shanghai"); err != nil {
		panic(err)
	}
}

func init() {
	// IANA Asia/Tokyo
	var err error
	if jst, err = time.LoadLocation("Asia/Tokyo"); err != nil {
		panic(err)
	}
}

// CST China Standard Time
func CST() *time.Location {
	return cst
}

// JST Japan Standard Time
func JST() *time.Location {
	return jst
}

// NowInCST return a time now in cst
func NowInCST() *time.Time {
	ts := time.Now().In(cst)
	return &ts
}

// NowInJST return a time now in jst
func NowInJST() *time.Time {
	ts := time.Now().In(jst)
	return &ts
}
