package server

import (
	"github.com/bluekaki/pkg/vv/internal/interceptor"
)

// Payload rest or grpc payload
type Payload = interceptor.Payload

// RegisteAuthorizationValidator some handler(s) for validate authorization and return userinfo
func RegisteAuthorizationValidator(name string, handler interceptor.UserinfoHandler) {
	interceptor.Validator.RegisteAuthorizationValidator(name, handler)
}

// RegisteProxyAuthorizationValidator some handler(s) for validate signature
func RegisteProxyAuthorizationValidator(name string, handler interceptor.SignatureHandler) {
	interceptor.Validator.RegisteProxyAuthorizationValidator(name, handler)
}
