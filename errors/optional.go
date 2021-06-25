package errors

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
)

type Enum interface {
	BZCode() int
	StatueCode() int
	Desc() string
}

func NewEnum(bzCode, statusCode int, desc string) Enum {
	bzErr := new(bzError)
	bzErr.code.bz = bzCode
	bzErr.code.status = statusCode
	bzErr.desc = desc

	return bzErr
}

type Error interface {
	error
	t()
}

var _ BzError = (*bzError)(nil)

type BzError interface {
	Error
	Enum
	Err() error
}

type bzError struct {
	code struct {
		bz     int
		status int
	}
	desc string
	err  *item
}

func (b *bzError) BZCode() int {
	return b.code.bz
}

func (b *bzError) StatueCode() int {
	return b.code.status
}

func (b *bzError) Desc() string {
	return b.desc
}

func (b *bzError) Err() error {
	return b.err
}

func (b *bzError) Error() string {
	if b.err == nil {
		return "nil"
	}
	return b.err.Error()
}

func (b *bzError) t() {}

func NewBzError(enum Enum, err error) Error {
	if enum == nil {
		return nil
	}

	bzErr := new(bzError)
	bzErr.code.bz = enum.BZCode()
	bzErr.code.status = enum.StatueCode()
	bzErr.desc = enum.Desc()

	if err != nil {
		if e, ok := err.(*item); ok {
			bzErr.err = e

		} else {
			bzErr.err = &item{msg: err.Error(), stack: callers(4)}
		}
	}

	return bzErr
}

var _ AlertError = (*alertError)(nil)

type AlertError interface {
	Error
	BzError() BzError
	AlertMessage() *AlertMessage
}

type alertError struct {
	bzError BzError
	alert   *AlertMessage
}

func (a *alertError) BzError() BzError {
	return a.bzError
}

func (a *alertError) AlertMessage() *AlertMessage {
	return a.alert
}

func (a *alertError) Error() string {
	err := a.bzError.Err()
	if err == nil {
		return "nil"
	}

	return err.Error()
}

func (a *alertError) t() {}

func NewAlertError(enum Enum, err error, projectName, journalID string, meta *AlertMessage_Meta) Error {
	if enum == nil {
		return nil
	}

	ts, _ := ptypes.TimestampProto(time.Now())
	alertErr := &alertError{
		bzError: NewBzError(enum, err).(BzError),
		alert: &AlertMessage{
			ProjectName: projectName,
			JournalId:   journalID,
			Meta:        meta,
			Ts:          ts,
		},
	}
	alertErr.alert.ErrorVerbose = fmt.Sprintf("%v", alertErr.bzError.Err())

	return alertErr
}
