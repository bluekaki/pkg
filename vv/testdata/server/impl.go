package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/vv"
	"github.com/bluekaki/pkg/vv/pkg/plugin/cuzerr"
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

var (
	bzCode    = cuzerr.NewCode(1101, http.StatusBadRequest, "some business error occurs")
	alertCode = cuzerr.NewCode(2307, http.StatusExpectationFailed, "some alert error occurs")
)

func (d *dummyService) Echo(ctx context.Context, req *dummy.EchoReq) (*dummy.EchoResp, error) {
	userinfo := vv.Userinfo(ctx).(*Userinfo)
	identifier := vv.SignatureIdentifier(ctx)

	if req.Message == "panic" {
		panic("a dummy panic err")
	}

	if req.Message == "business err" {
		return nil, cuzerr.NewBzError(bzCode, errors.New("got a business err"))
	}

	if req.Message == "alert err" {
		return nil, cuzerr.NewBzError(alertCode, errors.New("got an alert err")).AlertError(nil)
	}

	return &dummy.EchoResp{
		Message: fmt.Sprintf("Hello %s[%s], %s.", userinfo.Name, identifier, req.Message),
		Ack:     true,
	}, nil
}

func (d *dummyService) StreamEcho(stream dummy.DummyService_StreamEchoServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			if vv.IsValidatorError(err) {
				return err
			}
			if err != io.EOF {
				return cuzerr.NewBzError(alertCode, errors.Wrap(err, "stream recv err")).AlertError(nil)
			}
			break
		}

		err = stream.Send(&dummy.EchoResp{
			Message: fmt.Sprintf("Hello %s.", req.Message),
			Ack:     true,
		})
		if err != nil {
			return cuzerr.NewBzError(alertCode, errors.Wrap(err, "stream send err")).AlertError(nil)
		}
	}

	return nil
}
