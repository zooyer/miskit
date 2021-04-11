package zrpc

import (
	"context"
	"encoding/json"
	"github.com/zooyer/miskit/log"
	"github.com/zooyer/miskit/trace"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	var (
		ctx    = context.Background()
		api    = "https://apis.map.qq.com/ws/district/v1/list"
		params = NewForm(map[string]string{"key": "OB4BZ-D4W3U-B7VVO-4PJWW-6TKDJ-WPB77"})

		config = log.Config{
			Level: "DEBUG",
		}
		stdout    = log.NewStdoutRecorder(log.TextFormatter(true))
		logger, _ = log.New(config, nil)
		client    = New("test", 0, time.Second, logger)
	)
	logger.SetDefaultRecorder(stdout)

	traceInfo := trace.New(nil, "test")
	traceInfo.SpanID = "1234563"
	traceInfo.Tag = "test"
	traceInfo.Lang = "zh-CN"
	traceInfo.Content = json.RawMessage(`{"key":"val"}`)
	traceInfo.Request = &http.Request{
		URL: &url.URL{
			Scheme:      "",
			Opaque:      "",
			User:        nil,
			Host:        "",
			Path:        "/test/api/v1/call",
			RawPath:     "",
			ForceQuery:  false,
			RawQuery:    "",
			Fragment:    "",
			RawFragment: "",
		},
	}

	ctx = trace.Set(ctx, traceInfo)

	data, code, err := client.Get(ctx, api, params, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(code)
	t.Log(string(data))
}
