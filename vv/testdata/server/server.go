package main

import (
	"net"
	"time"

	"github.com/bluekaki/pkg/auth"
	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/shutdown"
	"github.com/bluekaki/pkg/vv/builder/server"
	"github.com/bluekaki/pkg/vv/proposal"
	"github.com/bluekaki/pkg/vv/testdata/api/gen"
	"github.com/bluekaki/pkg/zaplog"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var logger, _ = zaplog.NewJSONLogger()

func main() {
	defer logger.Sync()

	hcheck := health.NewServer()

	register := func(server *grpc.Server) {
		healthpb.RegisterHealthServer(server, hcheck)
		dummy.RegisterDummyServiceServer(server, &dummyService{logger: logger})
	}

	srv := server.New(logger, notifyHandler, register, server.WithProjectName("dummy-server"))

	listener, err := net.Listen("tcp", "127.0.0.1:8000")
	if err != nil {
		logger.Fatal("new grpc listener err", zap.Error(err))
	}

	go func() {
		logger.Info("server trying to listen on 127.0.0.1:8000")
		if err := srv.Serve(listener); err != nil {
			logger.Fatal("start server err", zap.Error(err))
		}
	}()

	shutdown.NewHook().Close(
		func() {
			srv.GracefulStop()
			hcheck.Shutdown()
			logger.Info("shutdown")
		},
	)
}

func notifyHandler(msg *proposal.AlertMessage) {
	logger.Error("notify", zap.Any("journal", msg))
}

func init() {
	server.RegisteAuthorizationValidator("dummy_sso", func(authorization string, payload proposal.Payload) (userinfo interface{}, err error) {
		if authorization == "cBmhBrwHZ0dM5DJy9TK1" {
			return &Userinfo{Name: "minami"}, nil
		}

		return nil, errors.New("illegal token")
	})
}

func init() {
	secrets := map[auth.Identifier]auth.Secret{
		"TESDUM": "9VbN+~_+8*,9WJ}#}^ZaoW)0=E>AaK",
	}

	signature, err := auth.NewSignature(auth.WithSecrets(secrets), auth.WithSHA256(), auth.WithTTL(time.Minute))
	if err != nil {
		panic(err)
	}

	server.RegisteAuthorizationProxyValidator("dummy_sign", func(proxyAuthorization string, payload proposal.Payload) (identifier string, ok bool, err error) {
		identifier, ok, err = signature.Verify(proxyAuthorization, payload.Date(), auth.ToMethod(payload.Method()), payload.URI(), payload.Body())
		return
	})
}
