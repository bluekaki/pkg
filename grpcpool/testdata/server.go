package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/bluekaki/pkg/grpcpool/testdata/pb"
	"github.com/bluekaki/pkg/zaplog"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type helloServer struct{}

func (h *helloServer) Unary(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	if false {
		delay := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100) + 1
		time.Sleep(time.Millisecond * time.Duration(delay))
	}

	return &pb.HelloReply{Message: message}, nil
}

func server() *cobra.Command {
	logger, err := zaplog.NewJSONLogger(zaplog.WithFileRotationP("/data/grpcpool/server.json"))
	if err != nil {
		panic(err)
	}

	var addr string

	cmd := &cobra.Command{
		Use:   "server",
		Short: "grpc server",
		Run: func(cmd *cobra.Command, args []string) {
			var kaep = keepalive.EnforcementPolicy{
				MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
				PermitWithoutStream: true,            // Allow pings even when there are no active streams
			}

			var kasp = keepalive.ServerParameters{
				MaxConnectionIdle:     15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
				MaxConnectionAge:      30 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
				MaxConnectionAgeGrace: 5 * time.Second,  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
				Time:                  5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
				Timeout:               1 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
			}

			listener, err := net.Listen("tcp", addr)
			if err != nil {
				logger.Fatal(fmt.Sprintf("listen %s err", addr), zap.Error(err))
			}

			srv := grpc.NewServer(
				grpc.KeepaliveEnforcementPolicy(kaep),
				grpc.KeepaliveParams(kasp),
			)

			pb.RegisterHelloServiceServer(srv, new(helloServer))

			logger.Info("trying to listen on " + addr)
			if err = srv.Serve(listener); err != nil {
				logger.Fatal("grpc server err", zap.Error(err))
			}
		},
	}

	cmd.Flags().StringVar(&addr, "addr", ":8888", "grpc server listen on addr")

	return cmd
}
