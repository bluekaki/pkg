package interceptor

import (
	"fmt"
	"sync"

	"github.com/bluekaki/pkg/vv/pkg/plugin/interceptor/options"
	"github.com/bluekaki/pkg/vv/proposal"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

const _ = grpc.SupportPackageIsVersion7

var fdLock sync.RWMutex

// RegisteAuthorizationValidator userinfo handler for interceptor options.authorization
func RegisteAuthorizationValidator(name string, handler proposal.UserinfoHandler) {
	fdLock.Lock()
	defer fdLock.Unlock()

	if _, ok := handlers.Authorization[name]; ok {
		panic(fmt.Sprintf("authorization validator: %s has exists", name))
	}

	handlers.Authorization[name] = handler
}

// RegisteAuthorizationProxyValidator signature handler for interceptor options.authorization_proxy
func RegisteAuthorizationProxyValidator(name string, handler proposal.SignatureHandler) {
	fdLock.Lock()
	defer fdLock.Unlock()

	if _, ok := handlers.AuthorizationProxy[name]; ok {
		panic(fmt.Sprintf("authorization-proxy validator: %s has exists", name))
	}

	handlers.AuthorizationProxy[name] = handler
}

// RegisteWhitelistingValidator whiteling handler for interceptor options.whitelisting
func RegisteWhitelistingValidator(name string, handler proposal.WhitelistingHandler) {
	fdLock.Lock()
	defer fdLock.Unlock()

	if _, ok := handlers.Whitelisting[name]; ok {
		panic(fmt.Sprintf("whitelisting validator: %s has exists", name))
	}

	handlers.Whitelisting[name] = handler
}

type Role int

const (
	Gateway Role = iota + 1
	Client
	Server
)

// ResloveFileDescriptor reslove options from FileDescriptor
func ResloveFileDescriptor(role Role) {
	fdLock.Lock()
	defer fdLock.Unlock()

	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		serivces := fd.Services()
		for i := 0; i < serivces.Len(); i++ {
			serivce := serivces.Get(i)
			if serviceHandler, _ := proto.GetExtension(serivce.Options(), options.E_ServiceHandler).(*options.ServiceHandler); serviceHandler != nil {
				handlers.Services[string(serivce.FullName())] = serviceHandler

				if role == Gateway {
					if serviceHandler.Whitelisting != nil && *serviceHandler.Whitelisting != "" {
						if _, ok := handlers.Whitelisting[*serviceHandler.Whitelisting]; !ok {
							panic(fmt.Sprintf("%s whitelisting validator: %s not found", string(serivce.FullName()), *serviceHandler.Whitelisting))
						}
					}

				} else if role == Server {
					if serviceHandler.Authorization != nil && *serviceHandler.Authorization != "" {
						if _, ok := handlers.Authorization[*serviceHandler.Authorization]; !ok {
							panic(fmt.Sprintf("%s authorization validator: %s not found", string(serivce.FullName()), *serviceHandler.Authorization))
						}
					}

					if serviceHandler.AuthorizationProxy != nil && *serviceHandler.AuthorizationProxy != "" {
						if _, ok := handlers.AuthorizationProxy[*serviceHandler.AuthorizationProxy]; !ok {
							panic(fmt.Sprintf("%s authorization-proxy validator: %s not found", string(serivce.FullName()), *serviceHandler.AuthorizationProxy))
						}
					}
				}
			}

			methods := serivce.Methods()
			for k := 0; k < methods.Len(); k++ {
				method := methods.Get(k)
				fullMethod := fmt.Sprintf("/%s/%s", serivce.FullName(), method.Name())

				if methodHandler, _ := proto.GetExtension(method.Options(), options.E_MethodHandler).(*options.MethodHandler); methodHandler != nil {
					handlers.Methods[fullMethod] = methodHandler

					if role == Gateway {
						if methodHandler.Whitelisting != nil && *methodHandler.Whitelisting != "" {
							if _, ok := handlers.Whitelisting[*methodHandler.Whitelisting]; !ok {
								panic(fmt.Sprintf("%s whitelisting validator: %s not found", fullMethod, *methodHandler.Whitelisting))
							}
						}

					} else if role == Server {
						if methodHandler.Authorization != nil && *methodHandler.Authorization != "" {
							if _, ok := handlers.Authorization[*methodHandler.Authorization]; !ok {
								panic(fmt.Sprintf("%s authorization validator: %s not found", fullMethod, *methodHandler.Authorization))
							}
						}

						if methodHandler.AuthorizationProxy != nil && *methodHandler.AuthorizationProxy != "" {
							if _, ok := handlers.AuthorizationProxy[*methodHandler.AuthorizationProxy]; !ok {
								panic(fmt.Sprintf("%s authorization-proxy validator: %s not found", fullMethod, *methodHandler.AuthorizationProxy))
							}
						}
					}
				}

				if httpRule, _ := proto.GetExtension(method.Options(), annotations.E_Http).(*annotations.HttpRule); httpRule != nil {
					handlers.HTTPRules[fullMethod] = httpRule
				}
			}
		}

		return true
	})
}

var handlers = &struct {
	Methods   map[string]*options.MethodHandler  // FullMethod : Handler
	Services  map[string]*options.ServiceHandler // FullMethod : Handler
	HTTPRules map[string]*annotations.HttpRule   // FullMethod : Rule

	Authorization      map[string]proposal.UserinfoHandler     // Name : Handler
	AuthorizationProxy map[string]proposal.SignatureHandler    // Name : Handler
	Whitelisting       map[string]proposal.WhitelistingHandler // Name : Handler
}{
	Methods:   make(map[string]*options.MethodHandler),
	Services:  make(map[string]*options.ServiceHandler),
	HTTPRules: make(map[string]*annotations.HttpRule),

	Authorization:      make(map[string]proposal.UserinfoHandler),
	AuthorizationProxy: make(map[string]proposal.SignatureHandler),
	Whitelisting:       make(map[string]proposal.WhitelistingHandler),
}

func FileDescriptorHasResloved(fullMethod string) bool {
	if _, ok := getMethodHandler(fullMethod); ok {
		return true
	}
	if _, ok := getServiceHandler(fullMethod); ok {
		return true
	}

	return false
}

func getMethodHandler(fullMethod string) (*options.MethodHandler, bool) {
	fdLock.RLock()
	defer fdLock.RUnlock()

	handler, ok := handlers.Methods[fullMethod]
	return handler, ok
}

func getServiceHandler(serviceName string) (*options.ServiceHandler, bool) {
	fdLock.RLock()
	defer fdLock.RUnlock()

	handler, ok := handlers.Services[serviceName]
	return handler, ok
}

func getHTTPRule(fullMethod string) (*annotations.HttpRule, bool) {
	fdLock.RLock()
	defer fdLock.RUnlock()

	rule, ok := handlers.HTTPRules[fullMethod]
	return rule, ok
}

func getAuthorizationHandler(name string) (proposal.UserinfoHandler, bool) {
	fdLock.RLock()
	defer fdLock.RUnlock()

	handler, ok := handlers.Authorization[name]
	return handler, ok
}

func getAuthorizationProxyHandler(name string) (proposal.SignatureHandler, bool) {
	fdLock.RLock()
	defer fdLock.RUnlock()

	handler, ok := handlers.AuthorizationProxy[name]
	return handler, ok
}

func getWhitelistingHandler(name string) (proposal.WhitelistingHandler, bool) {
	fdLock.RLock()
	defer fdLock.RUnlock()

	handler, ok := handlers.Whitelisting[name]
	return handler, ok
}
