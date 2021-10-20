package main

import (
	"context"
	"fmt"

	"github.com/bluekaki/pkg/vv"
	"github.com/bluekaki/pkg/vv/testdata/api/gen"

	"go.uber.org/zap"
)

type Userinfo struct {
	Name string
}

type dummyService struct {
	logger *zap.Logger
	dummy.UnimplementedDummyServiceServer
}

func (d *dummyService) Echo(ctx context.Context, req *dummy.EchoReq) (*dummy.EchoResp, error) {
	userinfo := vv.Userinfo(ctx).(*Userinfo)

	return &dummy.EchoResp{
		Message: fmt.Sprintf("Hello %s, %s.", userinfo.Name, req.Message),
		Ack:     true,
	}, nil
}
