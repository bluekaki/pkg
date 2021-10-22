package cuzerr

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/vv/proposal"
)

var _ proposal.Code = (*code)(nil)

type code struct {
	bzCode   int
	httpCode int
	desc     string
}

func (c *code) BzCode() int {
	return c.bzCode
}

func (c *code) HTTPCode() int {
	return c.httpCode
}

func (c *code) Desc() string {
	return c.desc
}

func (c *code) WithDesc(desc string) proposal.Code {
	clone := *c
	clone.desc = desc
	return &clone
}

var bzCodes = make(map[int]bool)

// NewCode create a proposal code
func NewCode(bzCode, httpCode int, desc string) proposal.Code {
	if bzCode <= 0 || bzCode > 99999999 {
		panic(fmt.Sprintf("bzCode %d illegal", bzCode))
	}
	if http.StatusText(httpCode) == "" {
		panic(fmt.Sprintf("httpCode %d not defined in http", httpCode))
	}

	if bzCodes[bzCode] {
		panic(fmt.Sprintf("bzCode %d duplicated", bzCode))
	}
	bzCodes[bzCode] = true

	return &code{
		bzCode:   bzCode,
		httpCode: httpCode,
		desc:     desc,
	}
}

type BzError interface {
	proposal.BzError
	AlertError(*proposal.AlertMessageMeta) proposal.AlertError
}

type bzError struct {
	proposal.Code
	err errors.Error
}

func (b *bzError) Error() string {
	return b.err.Error()
}

func (b *bzError) StackErr() errors.Error {
	return b.err
}

func (b *bzError) AlertError(meta *proposal.AlertMessageMeta) proposal.AlertError {
	return &alertError{
		bzError: b,
		alertMessage: &proposal.AlertMessage{
			ErrorVerbose: fmt.Sprintf("%+v", b.err),
			Meta:         meta,
			Timestamp:    time.Now(),
		},
	}
}

// NewBzError create a proposal bzerror
func NewBzError(code proposal.Code, err errors.Error) BzError {
	if code == nil {
		panic("proposal.Code required")
	}

	bzError := &bzError{err: err}
	bzError.Code = code

	return bzError
}

var _ proposal.AlertError = (*alertError)(nil)

type alertError struct {
	bzError      proposal.BzError
	alertMessage *proposal.AlertMessage
}

func (a *alertError) Error() string {
	return a.bzError.Error()
}

func (a *alertError) BzError() proposal.BzError {
	return a.bzError
}

func (a *alertError) AlertMessage() *proposal.AlertMessage {
	return a.alertMessage
}
