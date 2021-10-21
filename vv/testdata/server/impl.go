package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/vv"
	"github.com/bluekaki/pkg/vv/pkg/plugin/cuzerr"
	"github.com/bluekaki/pkg/vv/proposal"
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
	errorCode = cuzerr.NewCode(1101, http.StatusBadRequest, "some business error occurs")
	alertCode = cuzerr.NewCode(2307, http.StatusExpectationFailed, "some alert error occurs")
)

func (d *dummyService) Echo(ctx context.Context, req *dummy.EchoReq) (*dummy.EchoResp, error) {
	userinfo := vv.Userinfo(ctx).(*Userinfo)
	journalID, _ := vv.JournalID(ctx)

	if req.Message == "panic" {
		panic("a dummy panic err")
	}

	if req.Message == "business err" {
		return nil, cuzerr.NewBzError(
			errorCode,
			errors.New("got a business err"),
		)
	}

	if req.Message == "alert err" {
		return nil, cuzerr.NewAlertError(
			cuzerr.NewBzError(
				alertCode,
				errors.New("got an alert err"),
			),
			&proposal.AlertMessage{
				ProjectName:  "dummy-server",
				JournalID:    journalID,
				ErrorVerbose: fmt.Sprintf("%+v", errors.New("got an alert err")),
				Timestamp:    time.Now(),
			},
		)
	}

	return &dummy.EchoResp{
		Message: fmt.Sprintf("Hello %s, %s.", userinfo.Name, req.Message),
		Ack:     true,
	}, nil
}
