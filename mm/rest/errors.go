package rest

import (
	"encoding/json"
	"fmt"
)

var _ Error = (*mError)(nil)

// Error 包含返回码、描述语、错误的包装类
type Error interface {
	t()
	error
	// WithHTTPCode 设置返回的http code
	WithHTTPCode(httpCode int) Error
	// WithErr 设置真实发生的err，推荐使用github.com/pkg/errors包装一下stack信息，便于快速定位err发生的真实位置。
	WithErr(err error) Error
	// String 返回JSON格式的错误详情
	String() string
}

type mError struct {
	httpCode     int    // http码
	businessCode int    // 自定义的业务码
	desc         string // 错误描述语，用于接口返回
	err          error  // 实际发生错误，用于记录日志
}

// NewError 创建一个新的Error包装类
func NewError(businessCode int, desc string) Error {
	return &mError{
		businessCode: businessCode,
		desc:         desc,
	}
}

func (m *mError) t() {}

func (m *mError) Error() string {
	return fmt.Sprintf("[%d] %s", m.businessCode, m.desc)
}

func (m *mError) WithHTTPCode(httpCode int) Error {
	m.httpCode = httpCode
	return m
}

func (m *mError) WithErr(err error) Error {
	m.err = err
	return m
}

func (m *mError) String() string {
	err := &struct {
		HTTPCode     int    `json:"http_code"`
		BusinessCode int    `json:"business_code"`
		Desc         string `json:"desc"`
		Err          string `json:"err"`
	}{
		HTTPCode:     m.httpCode,
		BusinessCode: m.businessCode,
		Desc:         m.desc,
		Err:          fmt.Sprintf("%+v", m.err),
	}

	raw, _ := json.Marshal(err)
	return string(raw)
}

// ToError 尝试转换err
func ToError(err error) (Error, bool) {
	if err == nil {
		return nil, false
	}

	e, ok := err.(Error)
	return e, ok
}
