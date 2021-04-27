package timeutil

import (
	"time"

	"github.com/byepichi/pkg/errors"
)

// CSTLayout China Standard Time Layout
const CSTLayout = "2006-01-02 15:04:05"

// RFC3339ToCSTLayout convert rfc3339 value to china time layout
func RFC3339ToCSTLayout(value string) (string, error) {
	ts, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return "", errors.Wrapf(err, "time parse %s err", value)
	}

	return ts.In(cst).Format(CSTLayout), nil
}
