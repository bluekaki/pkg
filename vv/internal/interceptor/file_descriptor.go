package interceptor

import (
	"fmt"
	"sync"

	"github.com/bluekaki/pkg/vv/options"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const _ = grpc.SupportPackageIsVersion7

// FileDescriptor protobuf file descriptor
var FileDescriptor = &fileDescriptor{
	options: make(map[string]protoreflect.ProtoMessage),
}

type fileDescriptor struct {
	sync.RWMutex
	options map[string]protoreflect.ProtoMessage // FullMethod : Options
}

func (f *fileDescriptor) VerifyValidator(descriptor protoreflect.FileDescriptor) {
	f.Lock()
	defer f.Unlock()

	serivces := descriptor.Services()
	for i := 0; i < serivces.Len(); i++ {
		serivce := serivces.Get(i)
		f.options[string(serivce.FullName())] = serivce.Options()

		if option := proto.GetExtension(serivce.Options(), options.E_Authorization).(*options.Handler); option != nil &&
			Validator.AuthorizationValidator(option.Name) == nil {
			panic(fmt.Sprintf("%s options.authorization validator: [%s] not found", serivce.FullName(), option.Name))
		}

		if option := proto.GetExtension(serivce.Options(), options.E_ProxyAuthorization).(*options.Handler); option != nil &&
			Validator.ProxyAuthorizationValidator(option.Name) == nil {
			panic(fmt.Sprintf("%s options.proxy_authorization validator: [%s] not found", serivce.FullName(), option.Name))
		}
	}
}

func (f *fileDescriptor) ParseMethod(descriptor protoreflect.FileDescriptor, verifyValidator bool) {
	f.Lock()
	defer f.Unlock()

	serivces := descriptor.Services()
	for i := 0; i < serivces.Len(); i++ {
		serivce := serivces.Get(i)

		methods := serivce.Methods()
		for k := 0; k < methods.Len(); k++ {
			method := methods.Get(k)
			fullMethod := fmt.Sprintf("/%s/%s", serivce.FullName(), method.Name())
			f.options[fullMethod] = method.Options()

			if verifyValidator {
				if option := proto.GetExtension(method.Options(), options.E_MethodAuthorization).(*options.Handler); option != nil &&
					Validator.AuthorizationValidator(option.Name) == nil {
					panic(fmt.Sprintf("%s options.authorization validator: [%s] not found", fullMethod, option.Name))
				}

				if option := proto.GetExtension(method.Options(), options.E_MethodProxyAuthorization).(*options.Handler); option != nil &&
					Validator.ProxyAuthorizationValidator(option.Name) == nil {
					panic(fmt.Sprintf("%s options.proxy_authorization validator: [%s] not found", fullMethod, option.Name))
				}
			}
		}
	}
}

func (f *fileDescriptor) VerifyWhitelisting(descriptor protoreflect.FileDescriptor) {
	f.Lock()
	defer f.Unlock()

	serivces := descriptor.Services()
	for i := 0; i < serivces.Len(); i++ {
		serivce := serivces.Get(i)
		f.options[string(serivce.FullName())] = serivce.Options()

		if option := proto.GetExtension(serivce.Options(), options.E_Whitelisting).(*options.Handler); option != nil &&
			Validator.WhitelistingValidator(option.Name) == nil {
			panic(fmt.Sprintf("%s options.whitelisting validator: [%s] not found", serivce.FullName(), option.Name))
		}
	}

	for i := 0; i < serivces.Len(); i++ {
		serivce := serivces.Get(i)

		methods := serivce.Methods()
		for k := 0; k < methods.Len(); k++ {
			method := methods.Get(k)
			fullMethod := fmt.Sprintf("/%s/%s", serivce.FullName(), method.Name())
			f.options[fullMethod] = method.Options()

			if option := proto.GetExtension(method.Options(), options.E_MethodWhitelisting).(*options.Handler); option != nil &&
				Validator.WhitelistingValidator(option.Name) == nil {
				panic(fmt.Sprintf("%s options.whitelisting validator: [%s] not found", fullMethod, option.Name))
			}
		}
	}
}

func (f *fileDescriptor) Options(fullMethod string) protoreflect.ProtoMessage {
	f.RLock()
	defer f.RUnlock()

	return f.options[fullMethod]
}
