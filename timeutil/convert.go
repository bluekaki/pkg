package timeutil

import (
	"time"

	"github.com/bluekaki/pkg/errors"
)

const (
	// CSTLayout China Standard Time Layout
	CSTLayout = "2006-01-02 15:04:05"

	CSTMinuteLayout = "2006-01-02 15:04"

	CSTDayLayout = "2006-01-02"
)

// RFC3339ToCSTLayout convert rfc3339 value to china time layout
func RFC3339ToCSTLayout(value string) (string, error) {
	ts, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return "", errors.Wrapf(err, "time parse %s err", value)
	}

	return ts.In(cst).Format(CSTLayout), nil
}
