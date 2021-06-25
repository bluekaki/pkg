package interceptor

import (
	"sync"

	"github.com/bluekaki/pkg/errors"
)

type UserinfoHandler func(authorization string, payload Payload) (userinfo interface{}, err errors.Error)
type SignatureHandler func(proxyAuthorization string, payload Payload) (identifier string, ok bool, err errors.Error)
type WhitelistingHandler func(xForwardedFor string) (ok bool, err errors.Error)

// Validator authorization & proxy_authorization validator
var Validator = &validator{
	auth:         make(map[string]UserinfoHandler),
	proxyAuth:    make(map[string]SignatureHandler),
	whitelisting: make(map[string]WhitelistingHandler),
}

type validator struct {
	sync.RWMutex
	auth         map[string]UserinfoHandler
	proxyAuth    map[string]SignatureHandler
	whitelisting map[string]WhitelistingHandler
}

func (v *validator) RegisteAuthorizationValidator(name string, handler UserinfoHandler) {
	v.Lock()
	defer v.Unlock()

	v.auth[name] = handler
}

func (v *validator) RegisteProxyAuthorizationValidator(name string, handler SignatureHandler) {
	v.Lock()
	defer v.Unlock()

	v.proxyAuth[name] = handler
}

func (v *validator) RegisteWhitelistingValidator(name string, handler WhitelistingHandler) {
	v.Lock()
	defer v.Unlock()

	v.whitelisting[name] = handler
}

func (v *validator) AuthorizationValidator(name string) UserinfoHandler {
	v.RLock()
	defer v.RUnlock()

	return v.auth[name]
}

func (v *validator) ProxyAuthorizationValidator(name string) SignatureHandler {
	v.RLock()
	defer v.RUnlock()

	return v.proxyAuth[name]
}

func (v *validator) WhitelistingValidator(name string) WhitelistingHandler {
	v.RLock()
	defer v.RUnlock()

	return v.whitelisting[name]
}
