package sso

import (
	"context"
	"fmt"
	"github.com/zooyer/miskit/log"
	"github.com/zooyer/miskit/zrpc"
	"time"
)

type Option struct {
	AppID     int64
	AppSecret string
	Addr      string
	Retry     int
	Logger    *log.Logger
	Timeout   time.Duration
}

type Client struct {
	url    string
	option Option
	client *zrpc.Client
}

type Response struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

func New(option Option) *Client {
	return &Client{
		url:    fmt.Sprintf("http://%s/sso/api/v1/validate", option.Addr),
		option: option,
		client: zrpc.New("ice", option.Retry, option.Timeout, option.Logger),
	}
}

func (c *Client) LoginURL(ctx context.Context) string {
	return fmt.Sprintf("http://%s/sso/login", c.option.Addr)
}

func (c *Client) LogoutURL(ctx context.Context) string {
	return fmt.Sprintf("http://%s/sso/logout", c.option.Addr)
}

func (c *Client) LoggedURL(ctx context.Context) string {
	return fmt.Sprintf("http://%s/sso/logged", c.option.Addr)
}

func (c *Client) Validate(ctx context.Context, ticket string) (resp *Response, err error) {
	var request = map[string]interface{}{
		"app_id":     c.option.AppID,
		"app_secret": c.option.AppSecret,
		"ticket":     ticket,
	}
	var response Response

	if _, _, err = c.client.PostJSON(ctx, c.url, request, &response); err != nil {
		return
	}

	return &response, nil
}
