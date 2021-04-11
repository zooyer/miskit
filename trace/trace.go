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
	Caller  string          `json:"caller,omitempty"`
	TraceID string          `json:"trace_id,omitempty"`
	SpanID  string          `json:"span_id,omitempty"`
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

func New(req *http.Request, caller string) *Trace {
	var trace = new(Trace)

	trace.Caller = caller

	if req != nil {
		trace.TraceID = req.Header.Get(httpHeaderKeyTraceID)
		trace.SpanID = req.Header.Get(httpHeaderKeySpanID)
		trace.Lang = req.Header.Get(httpHeaderKeyLang)
		trace.Tag = req.Header.Get(httpHeaderKeyTag)
		trace.Content = []byte(req.Header.Get(httpHeaderKeyContent))
		trace.Request = req
	}

	if trace.TraceID == "" {
		trace.TraceID = GenTraceID()
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
	var trace = New(nil, "")
	switch ctx := ctx.(type) {
	case *gin.Context:
		trace, _ = ctx.MustGet(contextKey).(*Trace)
	default:
		trace, _ = ctx.Value(contextKey).(*Trace)
	}
	return trace
}

func (t *Trace) Set(header http.Header, caller string) {
	if t == nil {
		return
	}

	if t.TraceID != "" {
		header.Set(httpHeaderKeyTraceID, t.TraceID)
	} else {
		header.Set(httpHeaderKeyTraceID, GenTraceID())
	}
	header.Set(httpHeaderKeySpanID, GenSpanID())
	if t.Lang != "" {
		header.Set(httpHeaderKeyLang, t.Lang)
	}
	if t.Tag != "" {
		header.Set(httpHeaderKeyTag, t.Tag)
	}
	if caller != "" {
		header.Set(httpHeaderKeyCaller, caller)
	}
	if len(t.Content) > 0 {
		header.Set(httpHeaderKeyContent, string(t.Content))
	}
}

func (t *Trace) Child() *Trace {
	if t == nil {
		return nil
	}

	var trace Trace

	trace = *t
	trace.Content = make(json.RawMessage, len(t.Content))
	copy(trace.Content, t.Content)
	trace.SpanID = GenSpanID()

	return &trace
}

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

func (t *Trace) String() string {
	data, _ := json.Marshal(t)
	return string(data)
}

func GenTraceID() string {
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

func GenSpanID() string {
	return fmt.Sprintf("%x", rand.Int63())
}

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
