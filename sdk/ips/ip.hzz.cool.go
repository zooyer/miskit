package ips

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zooyer/miskit/log"
	"github.com/zooyer/miskit/zrpc"
	"time"
)

type Option struct {
	Retry   int
	Logger  *log.Logger
	Timeout time.Duration
}

type Client struct {
	url    string
	option Option
	client *zrpc.Client
}

type IP struct {
	IP       string `json:"ip"`
	Country  string `json:"country"`
	Province string `json:"province"`
	City     string `json:"city"`
	County   string `json:"county"`
	Region   string `json:"region"`
	ISP      string `json:"isp"`
}

func New(option Option) *Client {
	return &Client{
		url:    "https://ip.hzz.cool",
		option: option,
		client: zrpc.New("ips", option.Retry, option.Timeout, option.Logger),
	}
}

func (c *Client) QueryIP(ctx context.Context, ip string) (_ *IP, err error) {
	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data IP     `json:"data"`
	}

	var req = map[string]interface{}{
		"ip": ip,
	}

	data, _, err := c.client.Get(ctx, c.url, req, nil)
	if err != nil {
		return
	}

	if err = json.Unmarshal(data, &resp); err != nil {
		return
	}

	if resp.Code != 0 && resp.Code != 200 {
		return nil, fmt.Errorf("query ip fail, code: %d, msg:%s", resp.Code, resp.Msg)
	}

	return &resp.Data, nil
}
