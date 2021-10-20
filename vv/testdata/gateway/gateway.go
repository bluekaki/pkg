package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/bluekaki/pkg/ip"
	"github.com/bluekaki/pkg/shutdown"
	"github.com/bluekaki/pkg/vv/builder/gateway"
	"github.com/bluekaki/pkg/vv/proposal"
	"github.com/bluekaki/pkg/vv/testdata/api/gen"
	"github.com/bluekaki/pkg/zaplog"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var logger, _ = zaplog.NewJSONLogger()

func main() {
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())

	register := func(mux *runtime.ServeMux, opts []grpc.DialOption) error {
		return dummy.RegisterDummyServiceHandlerFromEndpoint(ctx, mux, "127.0.0.1:8000", opts)
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: gateway.NewCorsHandler(logger, notifyHandler, register, gateway.WithProjectName("dummy-gateway")),
	}

	go func() {
		logger.Info("gateway trying to listen on " + server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("start gateway err", zap.Error(err))
		}
	}()

	shutdown.NewHook().Close(
		func() {
			server.Shutdown(context.TODO())
			cancel()
			logger.Info("shutdown")
		},
	)
}

func notifyHandler(msg *proposal.AlertMessage) {
	logger.Error("notify", zap.Any("msg", msg))
}

func init() {
	filter, err := ip.NewFilter(
		&ip.Zone{Name: "internal", CIDR: []string{"10.0.0.0/8", "127.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}},
	)
	if err != nil {
		panic(err)
	}

	gateway.RegisteWhitelistingValidator("dummy_iplist", func(xForwardedFor string) (ok bool, err error) {
		forwarded := strings.Split(xForwardedFor, ",")
		realIP := strings.TrimSpace(forwarded[0])
		if realIP == "" {
			return false, nil
		}

		ok, _, err = filter.Bingo(realIP)
		return
	})
}
