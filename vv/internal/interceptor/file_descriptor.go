package interceptor

import (
	"fmt"

	"github.com/bluekaki/pkg/vv/pkg/plugin/interceptor/options"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

const _ = grpc.SupportPackageIsVersion7

func ResloveFileDescriptor() {
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		serivces := fd.Services()
		for i := 0; i < serivces.Len(); i++ {
			serivce := serivces.Get(i)
			if serviceHandler, _ := proto.GetExtension(serivce.Options(), options.E_ServiceHandler).(*options.ServiceHandler); serviceHandler != nil {
				handlers.Services[string(serivce.FullName())] = serviceHandler

				// TODO check
				fmt.Println("service:", string(serivce.FullName()), fmt.Sprintf("%+v", serviceHandler))
			}

			methods := serivce.Methods()
			for k := 0; k < methods.Len(); k++ {
				method := methods.Get(k)
				fullMethod := fmt.Sprintf("/%s/%s", serivce.FullName(), method.Name())

				if methodHandler, _ := proto.GetExtension(method.Options(), options.E_MethodHandler).(*options.MethodHandler); methodHandler != nil {
					handlers.Methods[fullMethod] = methodHandler

					// TODO check
					fmt.Println("method:", fullMethod, fmt.Sprintf("%+v", methodHandler))
				}

				if httpRule, _ := proto.GetExtension(method.Options(), annotations.E_Http).(*annotations.HttpRule); httpRule != nil {
					handlers.HttpRules[fullMethod] = httpRule

					// TODO check
					fmt.Println("httprule:", fullMethod, fmt.Sprintf("%+v", httpRule))
				}
			}
		}

		return true
	})
}

var handlers = &struct {
	Methods   map[string]*options.MethodHandler  // FullMethod : Handler
	Services  map[string]*options.ServiceHandler // FullMethod : Handler
	HttpRules map[string]*annotations.HttpRule   // FullMethod : Rule
}{
	Methods:   make(map[string]*options.MethodHandler),
	Services:  make(map[string]*options.ServiceHandler),
	HttpRules: make(map[string]*annotations.HttpRule),
}

func getMethodHandler(fullMethod string) (*options.MethodHandler, bool) {
	handler, ok := handlers.Methods[fullMethod]
	return handler, ok
}

func getServiceHandler(fullMethod string) (*options.ServiceHandler, bool) {
	handler, ok := handlers.Services[fullMethod]
	return handler, ok
}

func getHttpRule(fullMethod string) (*annotations.HttpRule, bool) {
	rule, ok := handlers.HttpRules[fullMethod]
	return rule, ok
}
