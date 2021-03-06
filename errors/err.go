package errors

import (
	"fmt"
	"io"
	"runtime"

	"github.com/pkg/errors"
)

// Error an error with stack info
type Error interface {
	error
	t()
}

func callers(skip int) []uintptr {
	var pcs [32]uintptr
	l := runtime.Callers(skip, pcs[:])
	return pcs[:l]
}

var _ (Error) = (*item)(nil)
var _ (fmt.Formatter) = (*item)(nil)

type item struct {
	msg   string
	stack []uintptr
}

func (i *item) Error() string {
	return i.msg
}

func (i *item) t() {}

// Format used by go.uber.org/zap in Verbose
func (i *item) Format(s fmt.State, verb rune) {
	io.WriteString(s, i.msg)
	io.WriteString(s, "\n")

	for _, pc := range i.stack {
		fmt.Fprintf(s, "%+v\n", errors.Frame(pc))
	}
}

// New create a new error
func New(msg string) Error {
	return &item{msg: msg, stack: callers(3)}
}

// Panic create a error from panic
func Panic(p interface{}) Error {
	return &item{msg: fmt.Sprintf("%+v", p), stack: callers(5)}
}

// Errorf create a new error
func Errorf(format string, args ...interface{}) Error {
	return &item{msg: fmt.Sprintf(format, args...), stack: callers(3)}
}

// Wrap with some extra message into err
func Wrap(err error, msg string) Error {
	if err == nil {
		return nil
	}

	e, ok := err.(*item)
	if !ok {
		return &item{msg: fmt.Sprintf("%s; %s", msg, err.Error()), stack: callers(3)}
	}

	e.msg = fmt.Sprintf("%s; %s", msg, e.msg)
	return e
}

// Wrapf with some extra message into err
func Wrapf(err error, format string, args ...interface{}) Error {
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf(format, args...)

	e, ok := err.(*item)
	if !ok {
		return &item{msg: fmt.Sprintf("%s; %s", msg, err.Error()), stack: callers(3)}
	}

	e.msg = fmt.Sprintf("%s; %s", msg, e.msg)
	return e
}

// WithStack add caller stack information
func WithStack(err error) Error {
	if err == nil {
		return nil
	}

	if e, ok := err.(*item); ok {
		return e
	}

	return &item{msg: err.Error(), stack: callers(3)}
}
