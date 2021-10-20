package proposal

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
	error
	Code
	StackErr() errors.Error
}

type AlertError interface {
	error
	BzError() BzError
	AlertMessage() *AlertMessage
}

type Validator interface {
	Validate() error
}

type NotifyHandler func(msg *AlertMessage)

type Signer func(fullMethod string, jsonRaw []byte) (authorizationProxy, date string, err error)

// Payload rest or grpc payload
type Payload interface {
	JournalID() string
	ForwardedByGrpcGateway() bool
	Service() string
	Date() string
	Method() string
	URI() string
	Body() []byte
}

type UserinfoHandler func(authorization string, payload Payload) (userinfo interface{}, err error)

type SignatureHandler func(authorizationProxy string, payload Payload) (identifier string, ok bool, err error)

type WhitelistingHandler func(xForwardedFor string) (ok bool, err error)
