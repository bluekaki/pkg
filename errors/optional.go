package errors

import (
	"fmt"
)

var _ BzError = (*bzError)(nil)

type Enum interface {
	BZCode() int
	StatueCode() int
	Desc() string
}

type BzError interface {
	Enum
	Err() Error
	t()
}

type bzError struct {
	code struct {
		bz     int
		status int
	}
	desc string
	err  *item
}

func (b *bzError) BZCode() int {
	return b.code.bz
}

func (b *bzError) StatueCode() int {
	return b.code.status
}

func (b *bzError) Desc() string {
	return b.desc
}

func (b *bzError) Err() Error {
	return b.err
}

func (b *bzError) t() {}

func NewBzError(enum Enum, err error) BzError {
	if enum == nil {
		return nil
	}

	bzErr := new(bzError)
	bzErr.code.bz = enum.BZCode()
	bzErr.code.status = enum.StatueCode()
	bzErr.desc = enum.Desc()

	if err != nil {
		if e, ok := err.(*item); ok {
			bzErr.err = e

		} else {
			bzErr.err = &item{msg: err.Error(), stack: callers(4)}
		}
	}

	return bzErr
}

var _ AlertError = (*alertError)(nil)

type AlertError interface {
	BzError() BzError
}

type alertError struct {
	bzError BzError
	alert   *AlertMessage
}

func (a *alertError) BzError() BzError {
	return a.bzError
}

func NewAlertError(enum Enum, err error, projectName, desc, journalID string) AlertError {
	if enum == nil {
		return nil
	}

	alertErr := &alertError{
		bzError: NewBzError(enum, err),
		alert: &AlertMessage{
			ProjectName: projectName,
			Desc:        desc,
			JournalId:   journalID,
		},
	}
	alertErr.alert.ErrorVerbose = fmt.Sprintf("%v", alertErr.bzError.Err())

	return alertErr
}
