package main

import (
	"context"
	"fmt"
	"math"
	"net/http"
	_ "net/http/pprof"
	"sync/atomic"
	"time"

	"github.com/byepichi/pkg/grpcpool"
	"github.com/byepichi/pkg/grpcpool/testdata/pb"
	"github.com/byepichi/pkg/zaplog"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func client() *cobra.Command {
	logger, err := zaplog.NewJSONLogger(zaplog.WithFileRotationP("/data/grpcpool/client.json"))
	if err != nil {
		panic(err)
	}

	var addr string
	var goroutines int
	var port int

	const maxRate float64 = 100000

	cmd := &cobra.Command{
		Use:   "client",
		Short: "grpc client",
		Run: func(cmd *cobra.Command, args []string) {
			limiter := rate.NewLimiter(rate.Limit(maxRate), 10000)
			go func() {
				for {
					for radian := float64(0); radian < 3.2; radian += 0.1 {
						limiter.SetLimit(rate.Limit(math.Sin(radian) * maxRate))
						time.Sleep(time.Second * 3)
					}
				}
			}()

			var kacp = keepalive.ClientParameters{
				Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
				Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
				PermitWithoutStream: true,             // send pings even without active streams
			}

			builder := func() (*grpc.ClientConn, error) {
				return grpc.Dial(addr,
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

			summary := uint64(0)
			call := func() {
				if !limiter.Allow() {
					time.Sleep(limiter.Reserve().Delay())
				}

				stub, err := pool.Get()
				if err != nil {
					logger.Error("pool get stub err", zap.Error(err))
					return
				}
				defer pool.Restore(stub)

				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				_, err = pb.NewHelloServiceClient(stub.Conn()).Unary(ctx, &pb.HelloRequest{Message: message})
				if err == grpc.ErrClientConnClosing {
					logger.Warn("call server's unary err", zap.Error(err))
					return
				}
				if err != nil {
					logger.Error("call server's unary err", zap.Error(err))
				}

				if x := atomic.AddUint64(&summary, 1); x%10000 == 0 {
					logger.Info("snapshot", zap.Uint64("summary", x))
				}
			}

			for k := 0; k < goroutines; k++ {
				go func() {
					for {
						call()
					}
				}()
			}

			http.Handle("/metrics", promhttp.Handler())
			logger.Sugar().Infof("metrics trying to listen on :%d", port)
			if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
				logger.Error("metrics start err", zap.Error(err))
			}
		},
	}

	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:8888", "grpc server addr")
	cmd.Flags().IntVar(&goroutines, "goroutines", 100, "concurrent goroutines")
	cmd.Flags().IntVar(&port, "port", 8889, "http port")

	return cmd
}
