package ice

import (
	"context"
	"fmt"
	"github.com/zooyer/miskit/log"
	"github.com/zooyer/miskit/zrpc"
	"github.com/zooyer/miskit/zuid"
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

type ID struct {
	zuid.ID
}

func New(option Option) *Client {
	return &Client{
		url:    fmt.Sprintf("http://%s/ice/v1/idgen?app_id=%v&app_secret=%v", option.Addr, option.AppID, option.AppSecret),
		option: option,
		client: zrpc.New("ice", option.Retry, option.Timeout, option.Logger),
	}
}

func (c *Client) GenID(ctx context.Context) (id ID, err error) {
	var response struct {
		ID int64 `json:"id"`
	}

	if _, _, err = c.client.Get(ctx, c.url, nil, &response); err != nil {
		return
	}

	return ID{
		ID: zuid.ID(response.ID),
	}, nil
}
