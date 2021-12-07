package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/vv"
	"github.com/bluekaki/pkg/vv/internal/pkg/multipart"
	"github.com/bluekaki/pkg/vv/pkg/plugin/cuzerr"
	"github.com/bluekaki/pkg/vv/testdata/api/gen"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (d *dummyService) Ping(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return new(emptypb.Empty), nil
}

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

	// panic("xxxxxxxxxxxxxxxxxxxxxxxxx")

	for k := 0; k < 3; k++ {
		err := stream.Send(&dummy.EchoResp{
			Message: fmt.Sprintf("{%s} %s[%s], %s #%d.", journalID, userinfo.Name, identifier, req.Message, k),
			Ack:     true,
		})
		if err != nil {
			return cuzerr.NewBzError(alertCode, errors.Wrap(err, "stream send err")).AlertError(nil)
		}

		// return cuzerr.NewBzError(alertCode, errors.New("stream alert err")).AlertError(nil)
	}

	return nil
}

func (d *dummyService) PostEcho(ctx context.Context, req *dummy.PostEchoReq) (*dummy.PostEchoResp, error) {

	return &dummy.PostEchoResp{
		Message: fmt.Sprintf("%s-%s", req.Name, req.Message),
		Ack:     true,
	}, nil
}

func (d *dummyService) Upload(ctx context.Context, req *dummy.UploadReq) (*dummy.UploadResp, error) {
	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	fmt.Println(req.FileName, len(req.Raw))
	fmt.Println("<<<<<<<<<<<<<<<<<<<<<<<<<<<<")

	digest := sha256.Sum256(bytes.Join(multipart.ParseFormData(req.Raw), nil))
	return &dummy.UploadResp{
		Digest: hex.EncodeToString(digest[:]),
	}, nil
}

func (d *dummyService) Picture(ctx context.Context, req *dummy.PictureReq) (*dummy.PictureResp, error) {
	raw, err := os.ReadFile(req.FileName)
	if err != nil {
		return nil, cuzerr.NewBzError(alertCode, errors.Wrapf(err, "read picture %s err", req.FileName)).AlertError(nil)
	}

	return &dummy.PictureResp{
		Raw: raw,
	}, nil
}

func (d *dummyService) Excel(ctx context.Context, _ *emptypb.Empty) (*dummy.ExcelResp, error) {
	raw, err := os.ReadFile("excel.xlsx")
	if err != nil {
		return nil, cuzerr.NewBzError(alertCode, errors.Wrap(err, "read excel.xlsx err")).AlertError(nil)
	}

	return &dummy.ExcelResp{
		Raw: raw,
	}, nil
}
