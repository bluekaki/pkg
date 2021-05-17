package pbutil

import (
	"time"

	"github.com/bluekaki/pkg/errors"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/protobuf/types/known/durationpb"
)

// NewDuration create protobuf duration by time.Duration
func NewDuration(d time.Duration) *durationpb.Duration {
	return ptypes.DurationProto(d)
}

// ParseDuration get time.Duration from protobuf duration
func ParseDuration(duration *durationpb.Duration) (time.Duration, error) {
	if duration == nil {
		return 0, errors.New("duration required")
	}

	d, err := ptypes.Duration(duration)
	if err != nil {
		return 0, errors.Wrap(err, "parse duration err")
	}

	return d, nil
}
