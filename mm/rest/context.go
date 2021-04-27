package rest

import (
	"bytes"
	stdctx "context"
	stderr "errors"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/byepichi/pkg/mm/internal/journal"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

var contextPool = &sync.Pool{
	New: func() interface{} {
		return new(context)
	},
}

func newContext(ctx *gin.Context) Context {
	context := contextPool.Get().(*context)
	context.ctx = ctx
	return context
}

func releaseContext(ctx Context) {
	c := ctx.(*context)
	c.ctx = nil
	contextPool.Put(c)
}

const (
	_Alias          = "_alias_"
	_PayloadName    = "_payload_"
	_AbortErrorName = "_abort_error_"
	_JournalName    = "_journal_"
	_LoggerName     = "_logger_"
	_BodyName       = "_body_"
	_SessionName    = "_session_"
)

// Journal 记录内部流转信息
type Journal = journal.T

var _ Context = (*context)(nil)

// Context 上下文、支持方法包装
type Context interface {
	init()

	// ShouldBindQuery 反序列化querystring
	// tag: `form:"xxx"` (注：不要写成query)
	ShouldBindQuery(obj interface{}) error

	// ShouldBindPostForm 反序列化postform(querystring会被忽略)
	// tag: `form:"xxx"`
	ShouldBindPostForm(obj interface{}) error

	// ShouldBindForm 同时反序列化querystring和postform;
	// 当querystring和postform存在相同字段时，postform优先使用。
	// tag: `form:"xxx"`
	ShouldBindForm(obj interface{}) error

	// ShouldBindJSON 反序列化postjson
	// tag: `json:"xxx"`
	ShouldBindJSON(obj interface{}) error

	// ShouldBindURI 反序列化path参数(如路由路径为 /userinfo/:name)
	// tag: `uri:"xxx"`
	ShouldBindURI(obj interface{}) error

	// Method 请求的method
	Method() string

	// Host 请求的host
	Host() string

	// Path 请求的路径(不附带querystring)
	Path() string

	// URI unescape后的uri
	URI() string

	// ContentType 请求的ContentType
	ContentType() string

	// Header clone一份请求的header
	Header() http.Header

	// Data 自定义返回数据
	Data(code int, contentType string, data []byte)

	// Redirect 重定向
	Redirect(code int, location string)

	// RequestContext 获取请求的context(当client关闭后，会自动canceled)
	RequestContext() stdctx.Context

	// FormFile 获取第一个出现的上传文件
	FormFile(name string) (*multipart.FileHeader, error)

	// GetHeader 从request中读取header
	GetHeader(key string) string

	// WriteHeader 向response中写入header
	WriteHeader(key, value string)

	// Param 获取path参数(如路由路径为 /userinfo/:name)
	// 推荐使用ShouldBindURI
	Param(key string) string

	// GetQuery 获取querystring参数
	// 推荐使用ShouldBindQuery
	GetQuery(key string) (string, bool)

	// GetPostForm 获取postform参数
	// 推荐使用ShouldBindPostForm
	GetPostForm(key string) (string, bool)

	// RawData 返回request.body
	RawData() []byte

	// Session 返回session对象
	Session() interface{}

	setSession(session interface{})

	// Cookie 获取cookie
	Cookie(name string) (*http.Cookie, error)

	// SetCookie 回写cookie
	SetCookie(cookie *http.Cookie) error

	// GetPayload 获取payload(可以是最终返回的payload，也可以是上一个handler处理后的payload)
	GetPayload() interface{}

	// SetPayload 设置payload(可以是最终返回的payload，也可以是本次处理、传递给下一个handler的payload)
	SetPayload(payload interface{})

	// Journal 获取内部流转信息对象
	Journal() Journal

	setJournal(journal Journal)

	// Logger 获取日志实例
	Logger() *zap.Logger

	setLogger(logger *zap.Logger)

	// AbortWithError 终止并处理错误
	AbortWithError(err Error)

	abortError() *mError

	disableJournal()

	doJournal() bool

	Alias() string

	setAlias(path string)
}

type context struct {
	ctx *gin.Context
}

func (c *context) init() {
	body, err := c.ctx.GetRawData()
	if err != nil {
		panic(err)
	}

	c.ctx.Set(_BodyName, body)                                   // cache body是为了journal使用
	c.ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body)) // re-construct req body
}

// ShouldBindQuery 反序列化querystring
// tag: `form:"xxx"` (注：不要写成query)
func (c *context) ShouldBindQuery(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.Query)
}

// ShouldBindPostForm 反序列化postform(querystring会被忽略)
// tag: `form:"xxx"`
func (c *context) ShouldBindPostForm(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.FormPost)
}

// ShouldBindForm 同时反序列化querystring和postform;
// 当querystring和postform存在相同字段时，postform优先使用。
// tag: `form:"xxx"`
func (c *context) ShouldBindForm(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.Form)
}

// ShouldBindJSON 反序列化postjson
// tag: `json:"xxx"`
func (c *context) ShouldBindJSON(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.JSON)
}

// ShouldBindURI 反序列化path参数(如路由路径为 /userinfo/:name)
// tag: `uri:"xxx"`
func (c *context) ShouldBindURI(obj interface{}) error {
	return c.ctx.ShouldBindUri(obj)
}

// Method 请求的method
func (c *context) Method() string {
	return c.ctx.Request.Method
}

// Host 请求的host
func (c *context) Host() string {
	return c.ctx.Request.Host
}

// Path 请求的路径(不附带querystring)
func (c *context) Path() string {
	return c.ctx.Request.URL.Path
}

// URI unescape后的uri
func (c *context) URI() string {
	uri, _ := url.QueryUnescape(c.ctx.Request.URL.RequestURI())
	return uri
}

// ContentType 请求的ContentType
func (c *context) ContentType() string {
	return c.ctx.ContentType()
}

// Header clone一份请求的header
func (c *context) Header() http.Header {
	header := c.ctx.Request.Header

	clone := make(http.Header, len(header))
	for k, v := range header {
		value := make([]string, len(v))
		copy(value, v)

		clone[k] = value
	}
	return clone
}

// Data 自定义返回数据
func (c *context) Data(code int, contentType string, data []byte) {
	c.ctx.Data(code, contentType, data)
}

// Redirect 重定向
func (c *context) Redirect(code int, location string) {
	c.ctx.Redirect(code, location)
}

// RequestContext 获取请求的context(当client关闭后，会自动canceled)
func (c *context) RequestContext() stdctx.Context {
	return c.ctx.Request.Context()
}

// FormFile 获取第一个出现的上传文件
func (c *context) FormFile(name string) (*multipart.FileHeader, error) {
	return c.ctx.FormFile(name)
}

// GetHeader 从request中读取header
func (c *context) GetHeader(key string) string {
	return c.ctx.GetHeader(key)
}

// WriteHeader 向response中写入header
func (c *context) WriteHeader(key, value string) {
	c.ctx.Header(key, value)
}

// Param 获取path参数(如路由路径为 /userinfo/:name)
// 推荐使用ShouldBindURI
func (c *context) Param(key string) string {
	return c.ctx.Param(key)
}

// GetQuery 获取querystring参数
// 推荐使用ShouldBindQuery
func (c *context) GetQuery(key string) (string, bool) {
	return c.ctx.GetQuery(key)
}

// GetPostForm 获取postform参数
// 推荐使用ShouldBindPostForm
func (c *context) GetPostForm(key string) (string, bool) {
	return c.ctx.GetPostForm(key)
}

// RawData 返回request.body
func (c *context) RawData() []byte {
	body, ok := c.ctx.Get(_BodyName)
	if !ok {
		return nil
	}

	return body.([]byte)
}

// Session 返回session对象
func (c *context) Session() interface{} {
	session, _ := c.ctx.Get(_SessionName)
	return session
}

func (c *context) setSession(session interface{}) {
	c.ctx.Set(_SessionName, session)
}

// Cookie 获取cookie
func (c *context) Cookie(name string) (*http.Cookie, error) {
	return c.ctx.Request.Cookie(name)
}

// SetCookie 回写cookie
func (c *context) SetCookie(cookie *http.Cookie) error {
	if cookie == nil {
		return stderr.New("cookie required")
	}

	http.SetCookie(c.ctx.Writer, cookie)
	return nil
}

// GetPayload 获取payload(可以是最终返回的payload，也可以是上一个handler处理后的payload)
func (c *context) GetPayload() interface{} {
	payload, _ := c.ctx.Get(_PayloadName)
	return payload
}

// SetPayload 设置payload(可以是最终返回的payload，也可以是本次处理、传递给下一个handler的payload)
func (c *context) SetPayload(payload interface{}) {
	c.ctx.Set(_PayloadName, payload)
}

// Journal 获取内部流转信息对象
func (c *context) Journal() Journal {
	j, ok := c.ctx.Get(_JournalName)
	if !ok || j == nil {
		return nil
	}

	return j.(Journal)
}

func (c *context) setJournal(journal Journal) {
	c.ctx.Set(_JournalName, journal)
}

// Logger 获取日志实例
func (c *context) Logger() *zap.Logger {
	logger, ok := c.ctx.Get(_LoggerName)
	if !ok {
		return nil
	}

	return logger.(*zap.Logger)
}

func (c *context) setLogger(logger *zap.Logger) {
	c.ctx.Set(_LoggerName, logger)
}

// AbortWithError 终止并处理错误
func (c *context) AbortWithError(err Error) {
	if err != nil {
		err := err.(*mError)

		httpCode := err.httpCode
		if httpCode == 0 {
			httpCode = http.StatusInternalServerError
		}
		c.ctx.AbortWithStatus(httpCode)
		c.ctx.Set(_AbortErrorName, err)
	}
}

func (c *context) abortError() *mError {
	err, ok := c.ctx.Get(_AbortErrorName)
	if !ok {
		return nil
	}

	return err.(*mError)
}

func (c *context) disableJournal() {
	c.setJournal(nil)
}

func (c *context) doJournal() bool {
	return c.Journal() != nil
}

func (c *context) Alias() string {
	path, ok := c.ctx.Get(_Alias)
	if !ok {
		return ""
	}

	return path.(string)
}

func (c *context) setAlias(path string) {
	if path = strings.TrimSpace(path); path != "" {
		c.ctx.Set(_Alias, path)
	}
}
