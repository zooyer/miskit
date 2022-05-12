package zrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	url2 "net/url"
	"reflect"
	"strings"
	"time"

	"github.com/zooyer/miskit/log"
	"github.com/zooyer/miskit/metric"
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
	Errno   int             `json:"errno"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// HTTP请求选项
type Option func(ctx context.Context, req *Request)

// 请求表单(适用于GET参数、POST表单等)
func NewForm(v interface{}) url2.Values {
	var values = make(url2.Values)
	switch val := v.(type) {
	case url2.Values:
		return val
	case map[string]interface{}:
		for k, v := range val {
			values.Set(k, fmt.Sprint(v))
		}
		return values
	case map[string]string:
		for k, v := range val {
			values.Set(k, v)
		}
		return values
	}

	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Map:
		for it := val.MapRange(); it.Next(); {
			key := fmt.Sprint(it.Key().Interface())
			val := fmt.Sprint(it.Value().Interface())
			values.Set(key, val)
		}
		return values
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			var field = val.Field(i)
			if tag := val.Type().Field(i).Tag; tag.Get("binding") == "required" || !field.IsZero() {
				var name = tag.Get("json")
				if name == "" || name == "-" {
					name = val.Type().Field(i).Name
				}
				values.Set(name, fmt.Sprint(field.Interface()))
			}
		}
	}

	return values
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
func (c *Client) Get(ctx context.Context, url string, params url2.Values, response interface{}, opts ...Option) (data []byte, code int, err error) {
	return c.do(ctx, "GET", "", url, params, response, opts...)
}

// POST请求
func (c *Client) post(ctx context.Context, url, contentType string, request, response interface{}, opts ...Option) (data []byte, code int, err error) {
	return c.do(ctx, "POST", contentType, url, request, response, opts...)
}

// POST表单请求
func (c *Client) PostForm(ctx context.Context, url string, values url2.Values, response interface{}, opts ...Option) (data []byte, code int, err error) {
	return c.post(ctx, url, "application/x-www-form-urlencoded", values, response, opts...)
}

// POST JSON请求
func (c *Client) PostJSON(ctx context.Context, url string, request, response interface{}, opts ...Option) (data []byte, code int, err error) {
	return c.post(ctx, url, "application/json", request, response, opts...)
}

// 创建HTTP请求
func (c *Client) newRequest(ctx context.Context, method, contentType, url string, request interface{}, opts ...Option) (req *http.Request, err error) {
	var body io.Reader

	if method != "GET" {
		content, _, _ := mime.ParseMediaType(contentType)
		switch content {
		case "application/json":
			data, err := json.Marshal(request)
			if err != nil {
				return nil, err
			}
			body = bytes.NewReader(data)
		case "application/x-www-form-urlencoded":
			switch form := request.(type) {
			case url2.Values:
				body = strings.NewReader(form.Encode())
			case map[string]string:
				var values = make(url2.Values)
				for key, val := range form {
					values.Set(key, val)
				}
				body = strings.NewReader(values.Encode())
			case map[string]interface{}:
				var values = make(url2.Values)
				for key, val := range form {
					values.Set(key, fmt.Sprint(val))
				}
				body = strings.NewReader(values.Encode())
			default:
				return nil, errors.New("not support content type and request type:" + contentType)
			}
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
		var query = req.URL.Query()
		switch form := request.(type) {
		case url2.Values:
			for key, values := range form {
				for _, val := range values {
					query.Add(key, val)
				}
			}
		case map[string]string:
			for key, val := range form {
				query.Add(key, val)
			}
		case map[string]interface{}:
			for key, val := range form {
				query.Add(key, fmt.Sprint(val))
			}
		default:
			return nil, errors.New("not support request type")
		}

		req.URL.RawQuery = query.Encode()
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
		child = trace.Get(ctx).GenChild()
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
		child.SetHeader(req.Header)
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
			return body, code, err
		}

		// 断言业务层errno
		if errno := res.Errno; errno != 0 {
			return body, code, fmt.Errorf("%s: http resonse code:%d, message:%s", c.name, code, res.Message)
		}

		if err = json.Unmarshal(res.Data, response); err != nil {
			return body, code, err
		}
	}

	return body, code, nil
}
