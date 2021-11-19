package main

import (
	"context"
	"fmt"
	"io"

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

	conn, err := client.NewConn("127.0.0.1:8000", logger, notifyHandler,
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

	defer conn.Close()

	dummySvc := dummy.NewDummyServiceClient(conn)

	if false {
		fmt.Println("---------------------- normal ----------------------------")

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

	if false {
		fmt.Println("---------------------- panic ----------------------------")

		ctx := metadata.AppendToOutgoingContext(context.TODO(), "Authorization", "cBmhBrwHZ0dM5DJy9TK1")
		var header metadata.MD

		resp, err := dummySvc.Echo(ctx, &dummy.EchoReq{
			Message: "panic",
		},
			grpc.Header(&header),
		)
		if err != nil {
			fmt.Println(header.Get("Journal-Id")[0], err)

		} else {
			fmt.Println(header.Get("Journal-Id")[0], resp.Message)
		}
	}

	if false {
		fmt.Println("---------------------- business err ----------------------------")

		ctx := metadata.AppendToOutgoingContext(context.TODO(), "Authorization", "cBmhBrwHZ0dM5DJy9TK1")
		var header metadata.MD

		resp, err := dummySvc.Echo(ctx, &dummy.EchoReq{
			Message: "business err",
		},
			grpc.Header(&header),
		)
		if err != nil {
			fmt.Println(header.Get("Journal-Id")[0], err)

		} else {
			fmt.Println(header.Get("Journal-Id")[0], resp.Message)
		}
	}

	if false {
		fmt.Println("---------------------- alert err ----------------------------")

		ctx := metadata.AppendToOutgoingContext(context.TODO(), "Authorization", "cBmhBrwHZ0dM5DJy9TK1")
		var header metadata.MD

		resp, err := dummySvc.Echo(ctx, &dummy.EchoReq{
			Message: "alert err",
		},
			grpc.Header(&header),
		)
		if err != nil {
			fmt.Println(header.Get("Journal-Id")[0], err)

		} else {
			fmt.Println(header.Get("Journal-Id")[0], resp.Message)
		}
	}

	if false {
		fmt.Println("---------------------- stream ----------------------------")

		ctx := metadata.AppendToOutgoingContext(context.TODO(), "Authorization", "cBmhBrwHZ0dM5DJy9TK1")

		req := &dummy.EchoReq{
			Message: "Hello World !",
		}

		stream, err := dummySvc.StreamEcho(ctx, req)
		if err != nil {
			panic(err)
		}

		header, _ := stream.Header()
		for {
			resp, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				panic(err)
			}

			fmt.Println(header.Get("Journal-Id")[0], resp)
		}
	}
}

func notifyHandler(msg *proposal.AlertMessage) {
	logger.Error("notify", zap.Any("journal", msg))
}
