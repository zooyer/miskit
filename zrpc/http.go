package zrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zooyer/miskit/metric"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/zooyer/miskit/log"
	"github.com/zooyer/miskit/trace"
)

// HTTP客户端
type Client struct {
	name    string
	retry   int
	timeout time.Duration
	logger  *log.Logger
	client  http.Client
	option  []Option
}

// HTTP请求
type Request http.Request

type Response struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// HTTP请求选项
type Option func(ctx context.Context, req *Request)

// 请求表单(适用于GET参数、POST表单等)
func NewForm(v interface{}) map[string]interface{} {
	var m = make(map[string]interface{})
	switch val := v.(type) {
	case map[string]interface{}:
		return val
	case map[string]string:
		for k, v := range val {
			m[k] = v
		}
		return m
	}

	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Map:
		for it := val.MapRange(); it.Next(); {
			m[fmt.Sprint(it.Key().Interface())] = fmt.Sprint(it.Value().Interface())
		}
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			var field = val.Field(i)
			if tag := val.Type().Field(i).Tag; tag.Get("binding") == "required" || !field.IsZero() {
				var name = tag.Get("json")
				if name == "" || name == "-" {
					name = val.Type().Field(i).Name
				}
				m[name] = field.Interface()
			}
		}
	}

	return m
}

// 创建HTTP客户端
func New(name string, retry int, timeout time.Duration, logger *log.Logger, opts ...Option) *Client {
	var connTimeout = timeout / 5
	if timeout != 0 {
		if connTimeout.Milliseconds() < 5 {
			connTimeout = 5 * time.Millisecond
		}
		timeout -= connTimeout

		if timeout <= 0 {
			panic("timeout must be greater than 5ms")
		}
	}

	return &Client{
		name:    name,
		retry:   retry,
		timeout: timeout,
		logger:  logger,
		option:  opts,
		client: http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:       connTimeout,
					KeepAlive:     5 * time.Second,
					FallbackDelay: 0,
				}).DialContext,
				MaxIdleConns:    50,
				IdleConnTimeout: 5 * time.Second,
			},
			Timeout: timeout,
		},
	}
}

// GET请求
func (c *Client) Get(ctx context.Context, url string, params map[string]interface{}, response interface{}, opts ...Option) (data []byte, code int, err error) {
	return c.do(ctx, "GET", "", url, params, response, opts...)
}

// POST请求
func (c *Client) post(ctx context.Context, url, contentType string, request, response interface{}, opts ...Option) (data []byte, code int, err error) {
	return c.do(ctx, "POST", contentType, url, request, response, opts...)
}

// POST表单请求
func (c *Client) PostForm(ctx context.Context, url string, params map[string]interface{}, response interface{}, opts ...Option) (data []byte, code int, err error) {
	return c.post(ctx, url, "application/x-www-form-urlencoded", params, response, opts...)
}

// POST JSON请求
func (c *Client) PostJSON(ctx context.Context, url string, request, response interface{}, opts ...Option) (data []byte, code int, err error) {
	return c.post(ctx, url, "application/json", request, response, opts...)
}

// 创建HTTP请求
func (c *Client) newRequest(ctx context.Context, method, contentType, url string, request interface{}, opts ...Option) (req *http.Request, err error) {
	var body io.Reader

	if method != "GET" {
		switch contentType {
		case "application/json":
			data, err := json.Marshal(request)
			if err != nil {
				return nil, err
			}
			body = bytes.NewReader(data)
		default:
			return nil, errors.New("not support content type:" + contentType)
		}
	}

	if req, err = http.NewRequestWithContext(ctx, method, url, body); err != nil {
		return
	}

	req.Header.Set("Content-Type", contentType)

	for _, opt := range append(c.option, opts...) {
		if opt != nil {
			opt(ctx, (*Request)(req))
		}
	}

	if method == "GET" && request != nil {
		params, ok := request.(map[string]interface{})
		if !ok {
			return nil, errors.New("request must is map[string]interface{} type")
		}
		var form = req.URL.Query()
		for key, val := range params {
			form.Set(key, fmt.Sprint(val))
		}
		req.URL.RawQuery = form.Encode()
	}

	return
}

func (c *Client) marshalJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// 执行HTTP请求
func (c *Client) do(ctx context.Context, method, contentType, url string, request, response interface{}, opts ...Option) (body []byte, code int, err error) {
	var (
		start = time.Now()
		retry int
		req   *http.Request
		resp  *http.Response
		child = trace.Get(ctx).Child()
	)

	defer func() {
		var (
			code    = code
			caller  = ""
			callee  = url
			latency = time.Since(start)
		)

		if req != nil {
			callee = req.URL.Path
		} else {
			callee = strings.TrimPrefix(callee, "http://")
			callee = strings.TrimPrefix(callee, "https://")
			if index := strings.Index(callee, "/"); index >= 0 {
				callee = callee[index:]
			}
		}

		if err != nil && (code == 0 || code == http.StatusOK) {
			code = 599
		}

		if child != nil && child.Request != nil {
			caller = child.Request.URL.Path
		}

		metric.Rpc("zrpc", caller, callee, code, latency, map[string]interface{}{
			"name": c.name,
		})

		if c.logger != nil {
			output := c.logger.Info
			if err != nil {
				output = c.logger.Error
			}

			c.logger.Tag(
				false,
				"rpc", "http",
				"name", c.name,
				"method", method,
				"latency", latency,
				"retry", retry,
			)
			if child != nil {
				c.logger.Tag(false, "cspan_id", child.SpanID)
			}
			if req != nil {
				c.logger.Tag(false, "url", req.URL.String())
			} else {
				c.logger.Tag(false, "url", url)
			}
			if request != nil {
				c.logger.Tag(false, "req", c.marshalJSON(request))
			}
			if bytes.ContainsAny(body, "\t\r\n") {
				if data, err := json.Marshal(json.RawMessage(body)); err == nil {
					body = data
				} else {
					body = bytes.ReplaceAll(body, []byte("\r"), nil)
					body = bytes.ReplaceAll(body, []byte("\n"), nil)
				}
			}
			if len(body) > 0 {
				c.logger.Tag(false, "resp", string(body))
			}
			if err != nil {
				c.logger.Tag(false, "error", err.Error())
			}

			output(ctx)
		}
	}()

	// 创建请求
	if req, err = c.newRequest(ctx, method, contentType, url, request, opts...); err != nil {
		return
	}

	// 设置trace
	if child != nil {
		child.Set(req.Header, child.Caller)
	}

	// 请求重试
	for i := 0; i < c.retry+1; i++ {
		if resp, err = c.client.Do(req); err == nil {
			break
		}
		retry++
	}
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// 断言HTTP响应
	if code = resp.StatusCode; code != http.StatusOK {
		return nil, code, fmt.Errorf("%s: http response code:%d, status:%s", c.name, resp.StatusCode, resp.Status)
	}

	// 读取响应
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	// 解析业务层响应
	if response != nil {
		// 解析body
		var res Response
		if err = json.Unmarshal(body, &res); err != nil {
			return
		}

		// 断言业务层code
		if code = res.Code; code != 0 {
			return nil, code, fmt.Errorf("%s: http resonse code:%d, message:%s", c.name, res.Code, res.Message)
		}

		if err = json.Unmarshal(res.Data, response); err != nil {
			return
		}
	}

	return body, resp.StatusCode, nil
}
