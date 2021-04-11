package ice

import (
	"context"
	"fmt"
	"time"

	"github.com/zooyer/jsons"
	"github.com/zooyer/miskit/log"
	"github.com/zooyer/miskit/zrpc"
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

func New(option Option) *Client {
	return &Client{
		url:    fmt.Sprintf("http://%s/upm/v1/auth", option.Addr),
		option: option,
		client: zrpc.New("upm", option.Retry, option.Timeout, option.Logger),
	}
}

func (c *Client) Auth(ctx context.Context, user int64, cond map[string]interface{}) (auth bool, args jsons.Raw, err error) {
	var resp struct {
		Auth bool      `json:"auth"`
		Args jsons.Raw `json:"args"`
	}

	for key, val := range cond {
		if _, ok := val.(string); !ok {
			cond[key] = fmt.Sprint(val)
		}
	}

	var req = map[string]interface{}{
		"app_id":     c.option.AppID,
		"app_secret": c.option.AppSecret,
		"user_id":    user,
		"perm_cond":  cond,
	}

	if _, _, err = c.client.PostJSON(ctx, c.url, req, &resp); err != nil {
		return
	}

	return resp.Auth, resp.Args, nil
}
