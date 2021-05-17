package gateway

import (
	"github.com/bluekaki/pkg/vv/internal/interceptor"
)

// RegisteWhitelistingValidator some handler(s) for whitelisting signature
func RegisteWhitelistingValidator(name string, handler func(xForwardedFor string) (ok bool, err error)) {
	interceptor.Validator.RegisteWhitelistingValidator(name, handler)
}
