package adapter

import (
	"github.com/bluekaki/pkg/errors"
)

type Code interface {
	BzCode() uint16
	HTTPCode() uint16
	Desc() string
	WithDesc(string) Code
}

type BzError interface {
	Code
	Error() errors.Error
}

type AlertError interface {
	BzError
	AlertMessage() *AlertMessage
}

type Validator interface {
	Validate() error
}

type NotifyHandler func(msg *AlertMessage)
