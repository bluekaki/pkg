package main

import (
	"time"

	"github.com/bluekaki/pkg/grpcpool"
	"github.com/bluekaki/pkg/zaplog"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func dummy() *cobra.Command {
	logger, err := zaplog.NewJSONLogger(zaplog.WithFileRotationP("/data/grpcpool/dummy.json"))
	if err != nil {
		panic(err)
	}

	cmd := &cobra.Command{
		Use:   "dummy",
		Short: "grpc dummy",
		Run: func(cmd *cobra.Command, args []string) {
			var kacp = keepalive.ClientParameters{
				Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
				Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
				PermitWithoutStream: true,             // send pings even without active streams
			}

			builder := func() (*grpc.ClientConn, error) {
				return grpc.Dial("127.0.0.1:8888",
					grpc.WithTimeout(time.Second*2),
					grpc.WithBlock(),
					grpc.WithInsecure(),
					grpc.WithKeepaliveParams(kacp),
				)
			}

			pool, err := grpcpool.NewPool(builder, grpcpool.WithEnablePrometheus())
			if err != nil {
				logger.Fatal("grpc new pool err", zap.Error(err))
			}

			conns := make([]grpcpool.Stub, 10)
			for i := range conns {
				conns[i], err = pool.Get()
				if err != nil {
					logger.Fatal("grpc new stub err", zap.Error(err))
				}
			}

			go func() {
				time.Sleep(time.Second * 2)
				for i := range conns {
					pool.Restore(conns[i])
				}
			}()

			pool.Close()
			pool.Close()
		},
	}

	return cmd
}
