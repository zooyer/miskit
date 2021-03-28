package zrpc

import (
	"context"
	"encoding/json"
	"github.com/zooyer/miskit/trace"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	var (
		ctx    = context.Background()
		url    = "https://apis.map.qq.com/ws/district/v1/list"
		params = NewForm(map[string]string{"key": "OB4BZ-D4W3U-B7VVO-4PJWW-6TKDJ-WPB77"})
		client = New("test", 0, time.Second, nil, nil)
	)

	traceInfo := trace.New(nil)
	traceInfo.SpanID = "1234563"
	traceInfo.Tag = "test"
	traceInfo.Lang = "zh-CN"
	traceInfo.Content = json.RawMessage(`{"key":"val"}`)

	ctx = trace.Set(ctx, traceInfo)

	data, code, err := client.Get(ctx, url, params, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(code)
	t.Log(string(data))
}
