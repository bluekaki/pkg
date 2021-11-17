package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
	journalID, _ := vv.JournalID(ctx)
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
		Message: fmt.Sprintf("{%s} %s[%s], %s.", journalID, userinfo.Name, identifier, req.Message),
		Ack:     true,
	}, nil
}

func (d *dummyService) StreamEcho(req *dummy.EchoReq, stream dummy.DummyService_StreamEchoServer) error {
	journalID, _ := vv.JournalID(stream.Context())
	userinfo := vv.Userinfo(stream.Context()).(*Userinfo)
	identifier := vv.SignatureIdentifier(stream.Context())

	for k := 0; k < 3; k++ {
		err := stream.Send(&dummy.EchoResp{
			Message: fmt.Sprintf("{%s} %s[%s], %s #%d.", journalID, userinfo.Name, identifier, req.Message, k),
			Ack:     true,
		})
		if err != nil {
			return cuzerr.NewBzError(alertCode, errors.Wrap(err, "stream send err")).AlertError(nil)
		}
	}

	return nil
}

func (d *dummyService) Upload(ctx context.Context, req *dummy.UploadReq) (*dummy.UploadResp, error) {
	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	fmt.Println(req.FileName, string(req.Raw))
	fmt.Println("<<<<<<<<<<<<<<<<<<<<<<<<<<<<")

	digest := sha256.Sum256(req.Raw)
	return &dummy.UploadResp{
		Digest: hex.EncodeToString(digest[:]),
	}, nil
}
