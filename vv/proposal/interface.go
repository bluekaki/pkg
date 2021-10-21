package proposal

import (
	"github.com/bluekaki/pkg/errors"
)

// Code some enum with business/http code, and some short descriptive information.
type Code interface {
	BzCode() int
	HTTPCode() int
	Desc() string
	WithDesc(string) Code
}

// BzError a business error, do not send alert.
type BzError interface {
	error
	Code
	StackErr() errors.Error
}

// AlertError a critical error, which will send an alert.
type AlertError interface {
	error
	BzError() BzError
	AlertMessage() *AlertMessage
}

// Validator the validator for protobuf message fields
type Validator interface {
	Validate() error
}

// NotifyHandler a handler for send alert
type NotifyHandler func(msg *AlertMessage)

// Signer a handler for do sign
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

// UserinfoHandler a handler for sso
type UserinfoHandler func(authorization string, payload Payload) (userinfo interface{}, err error)

// SignatureHandler a handler for verify signature
type SignatureHandler func(authorizationProxy string, payload Payload) (identifier string, ok bool, err error)

// WhitelistingHandler a handler for filter ip
type WhitelistingHandler func(xForwardedFor string) (ok bool, err error)
