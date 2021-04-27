package server

import (
	"github.com/byepichi/pkg/vv/internal/interceptor"
)

// Payload rest or grpc payload
type Payload = interceptor.Payload

// RegisteAuthorizationValidator some handler(s) for validate authorization and return userinfo
func RegisteAuthorizationValidator(name string, handler func(authorization string, payload Payload) (userinfo interface{}, err error)) {
	interceptor.Validator.RegisteAuthorizationValidator(name, handler)
}

// RegisteProxyAuthorizationValidator some handler(s) for validate signature
func RegisteProxyAuthorizationValidator(name string, handler func(proxyAuthorization string, payload Payload) (ok bool, err error)) {
	interceptor.Validator.RegisteProxyAuthorizationValidator(name, handler)
}
