package interceptor

import (
	"sync"
)

type userinfoHandler func(authorization string, payload Payload) (userinfo interface{}, err error)
type signatureHandler func(proxyAuthorization string, payload Payload) (identifier string, ok bool, err error)
type whitelistingHandler func(xForwardedFor string) (ok bool, err error)

// Validator authorization & proxy_authorization validator
var Validator = &validator{
	auth:         make(map[string]userinfoHandler),
	proxyAuth:    make(map[string]signatureHandler),
	whitelisting: make(map[string]whitelistingHandler),
}

type validator struct {
	sync.RWMutex
	auth         map[string]userinfoHandler
	proxyAuth    map[string]signatureHandler
	whitelisting map[string]whitelistingHandler
}

func (v *validator) RegisteAuthorizationValidator(name string, handler userinfoHandler) {
	v.Lock()
	defer v.Unlock()

	v.auth[name] = handler
}

func (v *validator) RegisteProxyAuthorizationValidator(name string, handler signatureHandler) {
	v.Lock()
	defer v.Unlock()

	v.proxyAuth[name] = handler
}

func (v *validator) RegisteWhitelistingValidator(name string, handler whitelistingHandler) {
	v.Lock()
	defer v.Unlock()

	v.whitelisting[name] = handler
}

func (v *validator) AuthorizationValidator(name string) userinfoHandler {
	v.RLock()
	defer v.RUnlock()

	return v.auth[name]
}

func (v *validator) ProxyAuthorizationValidator(name string) signatureHandler {
	v.RLock()
	defer v.RUnlock()

	return v.proxyAuth[name]
}

func (v *validator) WhitelistingValidator(name string) whitelistingHandler {
	v.RLock()
	defer v.RUnlock()

	return v.whitelisting[name]
}
