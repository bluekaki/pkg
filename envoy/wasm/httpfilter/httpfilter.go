package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	idgen "github.com/bluekaki/pkg/id"
	"github.com/bluekaki/pkg/timeutil"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

const (
	version      = "2.2.2"
	maxBodyBytes = 2 << 10
)

type Journal struct {
	ContextID uint32    `json:"context_id"`
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Request   struct {
		Method    string   `json:"method"`
		Scheme    string   `json:"scheme"`
		Authority string   `json:"authority"`
		Prefix    string   `json:"prefix"`
		Path      string   `json:"path"`
		Headers   []string `json:"headers"`
		Body      string   `json:"body"`
	} `json:"request"`
	Response struct {
		StatusCode int      `json:"status_code"`
		Status     string   `json:"status"`
		Headers    []string `json:"headers"`
		Body       string   `json:"body"`
	} `json:"response"`
	Success     bool    `json:"success"`
	CostSeconds float64 `json:"cost_seconds"`
}

func (j *Journal) Marshal() string {
	return fmt.Sprintf(`{"context_id":%d,"id":"%s","timestamp":"%s","request":{"method":"%s","scheme":"%s","authority":"%s","prefix":"%s","path":"%s","headers":[%s],"body":%s},"response":{"status_code":%d,"status":"%s","headers":[%s],"body":%s},"success":%v,"cost_seconds":%v}`,
		j.ContextID,
		j.ID,
		j.Timestamp.Format("2006-01-02 15:04:05"),
		j.Request.Method,
		j.Request.Scheme,
		j.Request.Authority,
		j.Request.Prefix,
		j.Request.Path,
		strings.Join(j.Request.Headers, ","),
		fmt.Sprintf("%q", j.Request.Body),
		j.Response.StatusCode,
		j.Response.Status,
		strings.Join(j.Response.Headers, ","),
		fmt.Sprintf("%q", j.Response.Body),
		j.Success,
		j.CostSeconds,
	)
}

func main() {
	time.Local = timeutil.CST()
	proxywasm.SetVMContext(&vmContext{})
}

type vmContext struct{}

func (v *vmContext) OnVMStart(vmConfigurationSize int) types.OnVMStartStatus {
	return types.OnVMStartStatusOK
}

func (v *vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{
		ContextID: contextID,
	}
}

type pluginContext struct {
	ContextID uint32
}

func (p *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	return types.OnPluginStartStatusOK
}

func (p *pluginContext) OnPluginDone() bool {
	return true
}

func (p *pluginContext) OnQueueReady(queueID uint32) {}

func (p *pluginContext) OnTick() {}

func (p *pluginContext) NewTcpContext(contextID uint32) types.TcpContext {

	return nil
}

func (p *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpContext{
		ContextID: contextID,
	}
}

type httpContext struct {
	ContextID        uint32
	RequestBodySize  int
	ResponseBodySize int
	Journal          *Journal
}

func (h *httpContext) reset() {
	h.RequestBodySize = 0
	h.ResponseBodySize = 0
	h.Journal = nil
}

func (h *httpContext) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	h.reset()

	id, _ := proxywasm.GetHttpRequestHeader("journal-id")
	if id = strings.TrimSpace(id); id == "" {
		id = idgen.JournalID()
		proxywasm.ReplaceHttpRequestHeader("journal-id", id)
	}

	tmp, err := proxywasm.GetHttpRequestHeaders()
	if err != nil {
		proxywasm.LogErrorf("【failed to get request headers: %v】", err)
	}

	journal := &Journal{ContextID: h.ContextID, ID: id, Timestamp: time.Now()}
	journal.Request.Headers = make([]string, 0, len(tmp)-5)

	for _, kv := range tmp {
		switch kv[0] {
		case ":method":
			journal.Request.Method = kv[1]
		case ":scheme":
			journal.Request.Scheme = kv[1]
		case ":authority":
			journal.Request.Authority = kv[1]
		case ":path":
			journal.Request.Path = QueryUnescape(kv[1])
		case "journal-id":
			// do nothing
		default:
			journal.Request.Headers = append(journal.Request.Headers, fmt.Sprintf("%q", kv[0]+": "+kv[1]))
		}
	}

	if net.ParseIP(journal.Request.Authority) != nil || journal.Request.Authority == "localhost" {
		h.reset()
		return types.ActionContinue
	}

	sort.Strings(journal.Request.Headers)
	h.Journal = journal

	return types.ActionContinue
}

func (h *httpContext) OnHttpRequestBody(bodySize int, endOfStream bool) types.Action {
	h.RequestBodySize += bodySize
	if !endOfStream {
		return types.ActionPause
	}

	if h.Journal == nil || h.RequestBodySize > maxBodyBytes {
		h.reset()
		return types.ActionContinue
	}

	body, err := proxywasm.GetHttpRequestBody(0, h.RequestBodySize)
	if err != nil {
		proxywasm.LogErrorf("【get http request body err %v】", err)
	}

	h.Journal.Request.Body = FormUnescape(body)
	return types.ActionContinue
}

func (h *httpContext) OnHttpRequestTrailers(numTrailers int) types.Action {

	return types.ActionContinue
}

func (h *httpContext) OnHttpResponseHeaders(numHeaders int, endOfStream bool) types.Action {
	proxywasm.ReplaceHttpResponseHeader("wasm-filter-version", version)

	if h.Journal == nil {
		return types.ActionContinue
	}

	// contentType, _ := proxywasm.GetHttpResponseHeader("content-type")
	// if !(strings.Contains(contentType, "application/x-www-form-urlencoded") ||
	// 	strings.Contains(contentType, "application/json") ||
	// 	strings.Contains(contentType, "application/grpc") ||
	// 	strings.Contains(contentType, "application/vnd.spring-boot.actuator")) {
	// 	h.reset()
	// 	return types.ActionContinue
	// }

	proxywasm.ReplaceHttpResponseHeader("journal-id", h.Journal.ID)

	tmp, err := proxywasm.GetHttpResponseHeaders()
	if err != nil {
		proxywasm.LogErrorf("【failed to get response headers: %v】", err)
	}

	h.Journal.Response.Headers = make([]string, 0, len(tmp)-2)

	for _, kv := range tmp {
		switch kv[0] {
		case ":status":
			statusCode, _ := strconv.Atoi(kv[1])
			h.Journal.Response.StatusCode = statusCode
			h.Journal.Response.Status = http.StatusText(statusCode)
			h.Journal.Success = statusCode == http.StatusOK
		case "x-forwarded-prefix":
			h.Journal.Request.Prefix = kv[1]
			if h.Journal.Request.Prefix != "/" {
				length := len(h.Journal.Request.Prefix)
				if string(h.Journal.Request.Prefix[length-1]) == "/" {
					length--
				}
				h.Journal.Request.Path = h.Journal.Request.Path[length:]
			}
		default:
			h.Journal.Response.Headers = append(h.Journal.Response.Headers, fmt.Sprintf("%q", kv[0]+": "+kv[1]))
		}
	}

	sort.Strings(h.Journal.Response.Headers)
	return types.ActionContinue
}

func (h *httpContext) OnHttpResponseBody(bodySize int, endOfStream bool) types.Action {
	h.ResponseBodySize += bodySize
	if !endOfStream {
		return types.ActionPause
	}

	// http1.1

	if h.Journal == nil || h.ResponseBodySize > maxBodyBytes {
		h.reset()
		return types.ActionContinue
	}

	h.RecordJournal()
	return types.ActionContinue
}

func (h *httpContext) OnHttpResponseTrailers(numTrailers int) types.Action {
	if h.Journal == nil || h.ResponseBodySize > maxBodyBytes {
		h.reset()
		return types.ActionContinue
	}

	// grpc

	h.RecordJournal()
	return types.ActionContinue
}

func (h *httpContext) OnHttpStreamDone() {
	if h.Journal != nil {
		h.RecordJournal()
	}
}

func (h *httpContext) RecordJournal() {
	body, err := proxywasm.GetHttpResponseBody(0, h.ResponseBodySize)
	if err != nil {
		proxywasm.LogErrorf("【get http response body err %v】", err)
	}

	h.Journal.Response.Body = FormUnescape(body)
	h.Journal.CostSeconds = time.Since(h.Journal.Timestamp).Seconds()

	headers := [][2]string{
		{":method", "POST"},
		{":authority", "localhost"},
		{":path", "/log"},
		{"accept", "*/*"},
	}

	_, err = proxywasm.DispatchHttpCall("self_http_cluster", headers, []byte(h.Journal.Marshal()), nil, 5000, func(numHeaders, bodySize, numTrailers int) {})
	if err != nil {
		proxywasm.LogErrorf(">>>>>> wasm send log err >>>>>>>  %v", err)
	}

	h.reset()
}

func QueryUnescape(uri string) string {
	decodedUri, err := url.QueryUnescape(uri)
	if err != nil {
		return uri
	}

	return decodedUri
}

func FormUnescape(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}

	body := string(raw)
	values, _ := url.ParseQuery(body)
	if body != values.Encode() {
		return body
	}

	return QueryUnescape(body)
}
