package neeko

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zooyer/miskit/zrpc"
)

type Client struct {
	url    string
	addr   string
	client *zrpc.Client
}

type SystemInfo struct {
	SystemID string `json:"system_id"`
	DeviceID string `json:"device_id"`
}

func New() *Client {
	return &Client{
		addr:   "127.0.0.1:54321",
		client: zrpc.New("neeko", 1, 50*time.Millisecond, nil),
	}
}

func (c *Client) api(version int, uri string) string {
	return fmt.Sprintf("http://%s/neeko/api/v%d/%s", c.addr, version, strings.TrimLeft(uri, "/"))
}

func (c *Client) SystemInfo(ctx context.Context) (info *SystemInfo, err error) {
	var resp SystemInfo

	if _, _, err = c.client.Get(ctx, c.api(1, "/system/info"), nil, &resp); err != nil {
		return
	}

	return &resp, nil
}
