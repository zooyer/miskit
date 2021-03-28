package zrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// 业务层响应解析
type Responder interface {
	New() Responder // 创建新解析器
	Code() int      // 业务层状态码
	Body() []byte   // 业务层数据
	Error() error   // 业务层错误
	Retry() bool    // 业务层重试
}

// HTTP客户端
type Client struct {
	name      string
	retry     int
	timeout   time.Duration
	logger    *log.Logger
	responder Responder
	client    http.Client
	option    []Option
}

// HTTP请求
type Request http.Request

// HTTP请求选项
type Option func(req *Request)

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
func New(name string, retry int, timeout time.Duration, logger *log.Logger, responder Responder, opts ...Option) *Client {
	return &Client{
		name:      name,
		retry:     retry,
		timeout:   timeout,
		logger:    logger,
		responder: responder,
		option:    opts,
		client: http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:       timeout,
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
func (c *Client) Post(ctx context.Context, url, contentType string, request, response interface{}, opts ...Option) (data []byte, code int, err error) {
	return c.do(ctx, "POST", contentType, url, request, response, opts...)
}

// POST表单请求
func (c *Client) PostForm(ctx context.Context, url string, params map[string]interface{}, response interface{}, opts ...Option) (data []byte, code int, err error) {
	return c.Post(ctx, url, "application/x-www-form-urlencoded", params, response, opts...)
}

// POST JSON请求
func (c *Client) PostJSON(ctx context.Context, url string, request, response interface{}, opts ...Option) (data []byte, code int, err error) {
	return c.Post(ctx, url, "application/json", request, response, opts...)
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

	if t := trace.Get(ctx); t != nil {
		t.Set(req.Header, c.name)
	}

	for _, opt := range append(c.option, opts...) {
		opt((*Request)(req))
	}

	if method == "GET" {
		params, ok := request.(map[string]interface{})
		if !ok {
			return nil, errors.New("request must is map[string]interface{} type")
		}
		var form = make(map[string][]string)
		for key, val := range params {
			form[key] = []string{fmt.Sprint(val)}
		}
		req.Form = form
		req.URL.RawQuery = req.Form.Encode()
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
		start     = time.Now()
		retry     int
		req       *http.Request
		resp      *http.Response
		responder Responder
	)

	defer func() {
		latency := time.Since(start)
		var path = url
		if req != nil {
			path = req.URL.Path
		} else {
			path = strings.TrimPrefix(path, "http://")
			path = strings.TrimPrefix(path, "https://")
			if index := strings.Index(path, "/"); index >= 0 {
				path = path[index:]
			}
		}

		log.I("TODO metric:", c.name, path, latency, err)

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
				if responder != nil {
					c.logger.Tag(false, "errno", responder.Code())
				}
				c.logger.Tag(false, "error", err.Error())
			}

			output(ctx)
		}
	}()

	// 创建请求
	if req, err = c.newRequest(ctx, method, contentType, url, request, opts...); err != nil {
		return
	}

	// 请求重试
retry:
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
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("%s: http response code:%d, status:%s", c.name, resp.StatusCode, resp.Status)
	}

	// 读取响应
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	// 没有业务层解析器, 直接返回
	if responder = c.responder; responder == nil {
		return body, resp.StatusCode, nil
	}

	// 解析业务层响应
	if err = json.Unmarshal(body, responder); err != nil {
		return
	}

	// 断言业务层响应
	if err = responder.Error(); err != nil {
		// 业务层出错, 可重试
		if responder.Retry() && retry < c.retry {
			goto retry
		}
		return
	}

	// 解析业务层数据
	if response != nil {
		if err = json.Unmarshal(responder.Body(), response); err != nil {
			return
		}
	}

	return body, resp.StatusCode, nil
}
