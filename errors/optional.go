package errors

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes"
)

type Enum interface {
	Combcode() uint32
	BZCode() uint16
	StatueCode() uint16
	Desc() string
	t()
}

var mapping = make(map[uint32]bool)

func NewEnum(bzCode, statusCode uint16, desc string) Enum {
	if bzCode > 999 {
		panic(fmt.Sprintf("bzCode %d illegal", bzCode))
	}
	if statusCode > 999 {
		panic(fmt.Sprintf("statusCode %d illegal", statusCode))
	}
	if http.StatusText(int(statusCode)) == "" {
		panic(fmt.Sprintf("statusCode %d not defined in http", statusCode))
	}

	bzErr := new(bzError)
	bzErr.code.bz = bzCode
	bzErr.code.status = statusCode
	bzErr.desc = desc

	code := bzErr.Combcode()
	if mapping[code] {
		panic(fmt.Sprintf("combcode %d duplicated", code))
	}
	mapping[code] = true

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
		bz     uint16
		status uint16
	}
	desc string
	err  *item
}

func (b *bzError) Combcode() uint32 {
	val, _ := strconv.ParseUint(fmt.Sprintf("%03d%03d", b.code.status, b.code.bz), 10, 32)
	return uint32(val)
}

func (b *bzError) BZCode() uint16 {
	return b.code.bz
}

func (b *bzError) StatueCode() uint16 {
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

func NewAlertError(enum Enum, err error, projectName string, meta *AlertMessage_Meta) Error {
	if enum == nil {
		return nil
	}

	if err != nil {
		if e, ok := err.(*item); ok {
			err = e

		} else {
			err = &item{msg: err.Error(), stack: callers(4)}
		}
	}

	ts, _ := ptypes.TimestampProto(time.Now())
	alertErr := &alertError{
		bzError: NewBzError(enum, err).(BzError),
		alert: &AlertMessage{
			ProjectName: projectName,
			Meta:        meta,
			Ts:          ts,
		},
	}
	alertErr.alert.ErrorVerbose = fmt.Sprintf("%v", alertErr.bzError.Err())

	return alertErr
}

func (a *AlertMessage) Init() *AlertMessage {
	a.Ts, _ = ptypes.TimestampProto(time.Now())
	return a
}
