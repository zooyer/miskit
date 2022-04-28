package trace

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	var trace = New(nil, "test")
	trace.Lang = "zh-CN"
	trace.Tag = "压测"
	trace.Content = json.RawMessage(`{"key":"value"}`)

	// HTTP请求传递
	req, err := http.NewRequest("GET", "http://www.baidu.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	trace.SetHeader(req.Header)

	// baidu
	trace = New(req, "baidu")
	log.Println(trace.String())

	// context上下文传递
	var ctx = context.Background()
	ctx = Set(ctx, trace)
	trace = Get(ctx)
	log.Println(trace.String())

	// 下游请求
	trace = trace.GenChild()
	trace.SetHeader(req.Header)
	trace = New(req, "child")
	log.Println(trace.String())
}
