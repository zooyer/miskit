package micro

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zooyer/miskit/errors"
	"github.com/zooyer/miskit/log"
	"github.com/zooyer/miskit/trace"
)

func Logger(logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		query := ctx.Request.URL.RawQuery

		ctx.Next()

		code := ctx.Writer.Status()
		latency := time.Since(start)

		output := logger.Error
		switch {
		case code >= http.StatusOK && code < http.StatusMultipleChoices:
			output = logger.Info
		case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
			output = logger.Info
		case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
			output = logger.Warning
		default:
			output = logger.Error
		}

		if t := trace.Get(ctx); t != nil {
			if t.TraceID != "" {
				logger.Tag(false, "trace_id", t.TraceID)
			}
			if t.SpanID != "" {
				logger.Tag(false, "span_id", t.SpanID)
			}
			if t.Tag != "" {
				logger.Tag(false, "tag", t.Tag)
			}
			if t.Lang != "" {
				logger.Tag(false, "lang", t.Lang)
			}
			if len(t.Content) > 0 {
				logger.Tag(false, "content", string(t.Content))
			}
		}

		logger.Tag(
			false,
			"ip", ctx.ClientIP(),
			"method", ctx.Request.Method,
			"path", path,
			"query", query,
			"code", code,
			"latency", latency,
		)

		// TODO response

		output(ctx)
	}
}

func Recovery(logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if e := recover(); e != nil {
				logger.Tag(false, "panic", e).Error(ctx, string(debug.Stack()))
				errors.New(errors.ServicePanic, fmt.Errorf("%v", e)).Metric()
			}
		}()
		ctx.Next()
	}
}

func Trace(caller string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		trace.Set(ctx, trace.New(ctx.Request, caller))
	}
}
