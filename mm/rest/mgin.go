package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/byepichi/pkg/mm/internal/journal"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	cors "github.com/rs/cors/wrapper/gin"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

// OnPanicNotify 发生panic时通知用
type OnPanicNotify func(ctx Context, err interface{}, stackInfo string)

// RecordMetrics 记录prometheus指标用
// 如果使用AliasForRecordMetrics配置了别名，uri将被替换为别名。
type RecordMetrics func(method, uri string, success bool, httpCode, businessCode int, costSeconds float64)

// HandlerFunc 逻辑处理handler定义
type HandlerFunc func(Context)

// DisableJournal 标识某些请求不记录journal
func DisableJournal(ctx Context) {
	ctx.disableJournal()
}

// AliasForRecordMetrics 对请求uri起个别名，用于prometheus记录指标。
// 如：Get /userinfo/:username 这样的uri，因为username会有非常多的情况，这样记录prometheus指标会非常的不有好。
func AliasForRecordMetrics(path string) HandlerFunc {
	return func(ctx Context) {
		ctx.setAlias(path)
	}
}

// ErrReply 发生错误时返回的内容定义
type ErrReply struct {
	Code int    `json:"code"`
	Desc string `json:"desc"`
}

// RouterGroup 包装gin的RouterGroup
type RouterGroup interface {
	Group(string, ...HandlerFunc) RouterGroup
	IRoutes
}

var _ IRoutes = (*router)(nil)

// IRoutes 包装gin的IRoutes
type IRoutes interface {
	Any(string, ...HandlerFunc)
	GET(string, ...HandlerFunc)
	POST(string, ...HandlerFunc)
	DELETE(string, ...HandlerFunc)
	PATCH(string, ...HandlerFunc)
	PUT(string, ...HandlerFunc)
	OPTIONS(string, ...HandlerFunc)
	HEAD(string, ...HandlerFunc)
	Static(relativePath, root string)
	StaticFS(relativePath string, fs http.FileSystem)
}

type router struct {
	group *gin.RouterGroup
}

func (r *router) Group(relativePath string, handlers ...HandlerFunc) RouterGroup {
	group := r.group.Group(relativePath, wrapHandlers(handlers...)...)
	return &router{group: group}
}

func (r *router) Any(relativePath string, handlers ...HandlerFunc) {
	r.group.Any(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) GET(relativePath string, handlers ...HandlerFunc) {
	r.group.GET(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) POST(relativePath string, handlers ...HandlerFunc) {
	r.group.POST(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) DELETE(relativePath string, handlers ...HandlerFunc) {
	r.group.DELETE(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) PATCH(relativePath string, handlers ...HandlerFunc) {
	r.group.PATCH(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) PUT(relativePath string, handlers ...HandlerFunc) {
	r.group.PUT(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) OPTIONS(relativePath string, handlers ...HandlerFunc) {
	r.group.OPTIONS(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) HEAD(relativePath string, handlers ...HandlerFunc) {
	r.group.HEAD(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) Static(relativePath, root string) {
	r.group.Static(relativePath, root)
}

func (r *router) StaticFS(relativePath string, fs http.FileSystem) {
	r.group.StaticFS(relativePath, fs)
}

func wrapHandlers(handlers ...HandlerFunc) []gin.HandlerFunc {
	funcs := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handler := handler
		funcs[i] = func(c *gin.Context) {
			ctx := newContext(c)
			defer releaseContext(ctx)

			handler(ctx)
		}
	}

	return funcs
}

// WrapSessionHandler 用来处理session或token的入口，在之后的handler中只需ctx.Session()即可。
// 如果handler内部出现错误，推荐返回rest.Error类型的错误以便自定义提示语，否则将默认返回"Internal Server Error"。
func WrapSessionHandler(handler func(Context) (session interface{}, err error)) HandlerFunc {
	return func(ctx Context) {
		session, err := handler(ctx)
		if err != nil {
			if e, ok := ToError(err); ok {
				ctx.AbortWithError(e)
			} else {
				ctx.AbortWithError(NewError(http.StatusInternalServerError, "Internal Server Error").WithHTTPCode(http.StatusInternalServerError).WithErr(err))
			}
			return
		}

		ctx.setSession(session)
	}
}

var _ Mux = (*mux)(nil)

// Mux http mux
type Mux interface {
	http.Handler
	Group(relativePath string, handlers ...HandlerFunc) RouterGroup
}

type mux struct {
	engine *gin.Engine
}

func (m *mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m.engine.ServeHTTP(w, req)
}

func (m *mux) Group(relativePath string, handlers ...HandlerFunc) RouterGroup {
	return &router{
		group: m.engine.Group(relativePath, wrapHandlers(handlers...)...),
	}
}

// Option 自定义选项
type Option func(*option)

type option struct {
	disablePProf      bool
	disablePrometheus bool
	panicNotify       OnPanicNotify
	recordMetrics     RecordMetrics
	enableCors        bool
	marshalJournal    bool
}

// WithDisablePProf 禁用pprof
func WithDisablePProf() Option {
	return func(opt *option) {
		opt.disablePProf = true
	}
}

// WithDisableproPrometheus 禁用prometheus
func WithDisableproPrometheus() Option {
	return func(opt *option) {
		opt.disablePrometheus = true
	}
}

// WithPanicNotify 设置panic时的通知回调
func WithPanicNotify(notify OnPanicNotify) Option {
	return func(opt *option) {
		opt.panicNotify = notify
	}
}

// WithRecordMetrics 设置记录prometheus记录指标回调
func WithRecordMetrics(recoder RecordMetrics) Option {
	return func(opt *option) {
		opt.recordMetrics = recoder
	}
}

// WithEnableCors 开启CORS
func WithEnableCors() Option {
	return func(opt *option) {
		opt.enableCors = true
	}
}

// WithMarshalJournal marshal journal to json string
func WithMarshalJournal() Option {
	return func(opt *option) {
		opt.marshalJournal = true
	}
}

// NewMux 创建http mux实例
func NewMux(logger *zap.Logger, options ...Option) (Mux, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}

	// withoutJournalPaths 这些请求，默认不记录journal
	withoutJournalPaths := map[string]bool{
		"/metrics": true,

		"/debug/pprof/":             true,
		"/debug/pprof/cmdline":      true,
		"/debug/pprof/profile":      true,
		"/debug/pprof/symbol":       true,
		"/debug/pprof/trace":        true,
		"/debug/pprof/allocs":       true,
		"/debug/pprof/block":        true,
		"/debug/pprof/goroutine":    true,
		"/debug/pprof/heap":         true,
		"/debug/pprof/mutex":        true,
		"/debug/pprof/threadcreate": true,

		"/favicon.ico": true,
	}

	opt := new(option)
	for _, f := range options {
		f(opt)
	}

	gin.DisableBindValidation()
	gin.SetMode(gin.ReleaseMode)
	mux := &mux{
		engine: gin.New(),
	}

	if !opt.disablePProf {
		pprof.Register(mux.engine) // registe pprof to gin
	}

	if !opt.disablePrometheus {
		mux.engine.GET("/metrics", gin.WrapH(promhttp.Handler())) // register prometheus
	}

	mux.engine.NoMethod(wrapHandlers(DisableJournal)...)
	mux.engine.NoRoute(wrapHandlers(DisableJournal)...)

	// recover两次，防止处理时发生panic，尤其是在OnPanicNotify中。
	mux.engine.Use(func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("got mgin panic", zap.String("panic", fmt.Sprintf("%+v", err)), zap.String("stack", string(debug.Stack())))
			}
		}()

		ctx.Next()
	})

	mux.engine.Use(func(ctx *gin.Context) {
		ts := time.Now()

		context := newContext(ctx)
		defer releaseContext(context)

		context.init()
		context.setLogger(logger)

		if !withoutJournalPaths[ctx.Request.URL.Path] {
			if journalID := context.GetHeader(journal.JournalHeader); journalID != "" {
				context.setJournal(journal.NewJournal(journalID))
			} else {
				context.setJournal(journal.NewJournal(""))
			}
		}

		defer func() {
			if err := recover(); err != nil {
				stackInfo := string(debug.Stack())
				context.AbortWithError(NewError(http.StatusInternalServerError, "Internal Server Error").WithHTTPCode(http.StatusInternalServerError).
					WithErr(fmt.Errorf("got application panic=> err: %v  stack: %s", err, stackInfo)))

				if notify := opt.panicNotify; notify != nil {
					notify(context, err, stackInfo)
				}
			}

			if x := context.Journal(); x != nil {
				context.WriteHeader(journal.JournalHeader, x.ID())
			}

			var (
				response interface{} // the resp body
				abortErr error
			)

			if ctx.IsAborted() {
				for i := range ctx.Errors { // gin error
					multierr.AppendInto(&abortErr, ctx.Errors[i])
				}

				if err := context.abortError(); err != nil { // customer err
					multierr.AppendInto(&abortErr, err.err)

					reply := &ErrReply{
						Code: err.businessCode,
						Desc: err.desc,
					}

					response = reply
					ctx.PureJSON(err.httpCode, reply)
				}

			} else {
				response = context.GetPayload()
				if response != nil {
					ctx.PureJSON(http.StatusOK, response)
				}
			}

			if ctx.Writer.Status() == http.StatusNotFound {
				return
			}

			if opt.recordMetrics != nil {
				uri := context.URI()
				if alias := context.Alias(); alias != "" {
					uri = alias
				}

				businessCode := 0
				if response != nil {
					if reply, ok := response.(*ErrReply); ok {
						businessCode = reply.Code
					}
				}

				opt.recordMetrics(context.Method(), uri, !ctx.IsAborted() && ctx.Writer.Status() == http.StatusOK, ctx.Writer.Status(), businessCode, time.Since(ts).Seconds())
			}

			var j *journal.Journal
			if x := context.Journal(); x != nil {
				j = x.(*journal.Journal)
			} else {
				return
			}

			j.WithRequest(&journal.Request{
				TTL:        "unlimit",
				Method:     ctx.Request.Method,
				DecodedURL: context.URI(),
				Header:     journal.ToJournalHeader(ctx.Request.Header),
				Body:       string(context.RawData()),
			})

			j.WithResponse(&journal.Response{
				Header:     journal.ToJournalHeader(ctx.Writer.Header()),
				StatusCode: ctx.Writer.Status(),
				Status:     http.StatusText(ctx.Writer.Status()),
				Body:       response,
			})
			j.Success = !ctx.IsAborted() && ctx.Writer.Status() == http.StatusOK
			j.CostSeconds = time.Since(ts).Seconds()

			var journal interface{} = j
			if opt.marshalJournal {
				raw, _ := json.Marshal(j)
				journal = string(raw)
			}

			if abortErr == nil {
				logger.Info("mgin interceptor", zap.Any("journal", journal))

			} else {
				logger.Error("mgin interceptor", zap.Any("journal", journal), zap.Error(abortErr))
			}
		}()

		ctx.Next()
	})

	if opt.enableCors {
		mux.engine.Use(cors.New(cors.Options{ // similar to cors.AllowAll()
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
			},
			AllowedHeaders:     []string{"*"},
			AllowCredentials:   true,
			OptionsPassthrough: true, // recommend; for preflight
		}))
	}

	return mux, nil
}
