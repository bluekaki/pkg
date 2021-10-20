package main

import (
	"context"
	"fmt"

	"github.com/bluekaki/pkg/auth"
	"github.com/bluekaki/pkg/vv/builder/client"
	"github.com/bluekaki/pkg/vv/proposal"
	"github.com/bluekaki/pkg/vv/testdata/api/gen"
	"github.com/bluekaki/pkg/zaplog"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var logger, _ = zaplog.NewJSONLogger()

func main() {
	defer logger.Sync()

	signer, err := auth.NewSignature(auth.WithSHA256(), auth.WithSecrets(
		map[auth.Identifier]auth.Secret{
			"TESDUM": "9VbN+~_+8*,9WJ}#}^ZaoW)0=E>AaK",
		}),
	)
	if err != nil {
		panic(err)
	}

	conn, err := client.New("127.0.0.1:8000", logger, notifyHandler,
		client.WithSigner(
			func(fullMethod string, jsonRaw []byte) (authorizationProxy, date string, err error) {
				return signer.Generate("TESDUM", auth.MethodGRPC, fullMethod, jsonRaw)
			},
		),
		client.WithProjectName("dummy-client"),
	)

	if err != nil {
		panic(err)
	}

	dummySvc := dummy.NewDummyServiceClient(conn)

	ctx := metadata.AppendToOutgoingContext(context.TODO(), "Authorization", "cBmhBrwHZ0dM5DJy9TK1")
	var header metadata.MD

	resp, err := dummySvc.Echo(ctx, &dummy.EchoReq{
		Message: "Hello World !",
	},
		grpc.Header(&header),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(header.Get("Journal-Id")[0], resp.Message)
}

func notifyHandler(msg *proposal.AlertMessage) {
	logger.Error("notify", zap.Any("msg", msg))
}
