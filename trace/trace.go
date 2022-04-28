package trace

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Trace struct {
	Callee  string          `json:"callee,omitempty"`
	Caller  string          `json:"caller,omitempty"`
	TraceID string          `json:"trace_id,omitempty"`
	SpanID  string          `json:"span_id,omitempty"`
	ChildID string          `json:"child_id,omitempty"`
	Lang    string          `json:"lang,omitempty"`
	Tag     string          `json:"tag,omitempty"`
	Content json.RawMessage `json:"content,omitempty"`
	Request *http.Request   `json:"-"`
}

const (
	httpHeaderKeyTraceID = "Z-TraceID"
	httpHeaderKeySpanID  = "Z-SpanID"
	httpHeaderKeyLang    = "Z-Lang"
	httpHeaderKeyTag     = "Z-Tag"
	httpHeaderKeyCaller  = "Z-Caller"
	httpHeaderKeyContent = "Z-Content"
)

const (
	contextKey = "z-context"
)

var (
	pool = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
)

func New(req *http.Request, callee string) *Trace {
	var trace = new(Trace)

	trace.Callee = callee

	if req != nil {
		trace.TraceID = req.Header.Get(httpHeaderKeyTraceID)
		trace.Caller = req.Header.Get(httpHeaderKeyCaller)
		trace.SpanID = req.Header.Get(httpHeaderKeySpanID)
		trace.Lang = req.Header.Get(httpHeaderKeyLang)
		trace.Tag = req.Header.Get(httpHeaderKeyTag)
		trace.Content = []byte(req.Header.Get(httpHeaderKeyContent))
		trace.Request = req
	}

	if trace.TraceID == "" {
		trace.TraceID = genTraceID()
	}

	return trace
}

func Set(ctx context.Context, trace *Trace) context.Context {
	if trace == nil {
		return ctx
	}

	switch ctx := ctx.(type) {
	case *gin.Context:
		ctx.Set(contextKey, trace)
	default:
		return context.WithValue(ctx, contextKey, trace)
	}

	return ctx
}

func Get(ctx context.Context) *Trace {
	var trace *Trace
	switch ctx := ctx.(type) {
	case *gin.Context:
		if value, exists := ctx.Get(contextKey); exists {
			trace, _ = value.(*Trace)
		}
	default:
		trace, _ = ctx.Value(contextKey).(*Trace)
	}
	return trace
}

// SetHeader 设置到请求头
func (t *Trace) SetHeader(header http.Header) {
	if t == nil {
		return
	}

	if t.TraceID != "" {
		header.Set(httpHeaderKeyTraceID, t.TraceID)
	} else {
		header.Set(httpHeaderKeyTraceID, genTraceID())
	}
	header.Set(httpHeaderKeySpanID, genSpanID())
	if t.Lang != "" {
		header.Set(httpHeaderKeyLang, t.Lang)
	}
	if t.Tag != "" {
		header.Set(httpHeaderKeyTag, t.Tag)
	}
	if t.Callee != "" {
		header.Set(httpHeaderKeyCaller, t.Callee)
	}
	if len(t.Content) > 0 {
		header.Set(httpHeaderKeyContent, string(t.Content))
	}
}

// GenChild 生成子trace
func (t *Trace) GenChild() *Trace {
	if t == nil {
		return nil
	}

	var trace Trace

	trace = *t
	trace.Content = make(json.RawMessage, len(t.Content))
	copy(trace.Content, t.Content)
	trace.SpanID = genSpanID()

	return &trace
}

// Clone 深度拷贝克隆
func (t *Trace) Clone() *Trace {
	if t == nil {
		return nil
	}

	var trace Trace

	trace = *t
	trace.Content = make(json.RawMessage, len(t.Content))
	copy(trace.Content, t.Content)

	return &trace
}

// String 序列化成字符串
func (t *Trace) String() string {
	data, _ := json.Marshal(t)
	return string(data)
}

// genTraceID 生成trace id
func genTraceID() string {
	pid := os.Getegid()
	now := time.Now()
	unix := now.Unix()
	nano := now.UnixNano()

	var buf = pool.Get().(*bytes.Buffer)
	defer pool.Put(buf)
	defer buf.Reset()

	buf.WriteString(hex.EncodeToString(net.ParseIP(getLocalIP()).To4()))
	buf.WriteString(fmt.Sprintf("%x", unix&0xffffffff))
	buf.WriteString(fmt.Sprintf("%04x", nano&0xffff))
	buf.WriteString(fmt.Sprintf("%04x", pid&0xffff))
	buf.WriteString(fmt.Sprintf("%06x", rand.Int31n(1<<24)))
	buf.WriteString("5a")

	return buf.String()
}

// genSpanID 生成
func genSpanID() string {
	return fmt.Sprintf("%x", rand.Int63())
}

// getLocalIP 获取本机IP地址
func getLocalIP() string {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		return "0.0.0.0"
	}

	for _, addr := range addr {
		if ip, ok := addr.(*net.IPNet); ok && ip.IP.IsGlobalUnicast() {
			return ip.IP.String()
		}
	}

	return "127.0.0.1"
}
