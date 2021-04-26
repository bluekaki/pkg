package pbutil

import (
	"time"

	"github.com/byepichi/pkg/errors"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// NewTimestamp create protobuf timestamp by time.Time
func NewTimestamp(ts *time.Time) (*timestamppb.Timestamp, error) {
	if ts == nil {
		return nil, errors.New("ts required")
	}

	timestamp, err := ptypes.TimestampProto(*ts)
	if err != nil {
		return nil, errors.Wrap(err, "new timestamp by ts err")
	}

	return timestamp, nil
}

// ParseTimestamp get time.Time from protobuf timestamp
func ParseTimestamp(timestamp *timestamppb.Timestamp, location ...*time.Location) (*time.Time, error) {
	if timestamp == nil {
		return nil, errors.New("timestamp required")
	}

	ts, err := ptypes.Timestamp(timestamp)
	if err != nil {
		return nil, errors.Wrap(err, "get ts from timestamp err")
	}

	var val time.Time
	if location == nil {
		val = ts.In(time.Local)
	} else {
		val = ts.In(location[0])
	}

	return &val, nil
}
