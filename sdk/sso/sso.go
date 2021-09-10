package sso

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zooyer/miskit/log"
	"github.com/zooyer/miskit/zrpc"
	"os/exec"
	"strings"
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

func (c *Client) Validate(ctx *gin.Context, ticket string) (resp *Response, err error) {
	var request = map[string]interface{}{
		"app_id":     c.option.AppID,
		"app_secret": c.option.AppSecret,
		"ticket":     ticket,
	}
	var response Response

	var header = []string{"Client-ID", "Client-UA", "Client-Info"}

	var option zrpc.Option = func(_ context.Context, req *zrpc.Request) {
		for _, key := range header {
			req.Header.Set(key, ctx.GetHeader(key))
		}
	}

	if _, _, err = c.client.PostJSON(ctx, c.url, request, &response, option); err != nil {
		return
	}

	return &response, nil
}

// GetCPUID 获取cpuid
func GetCPUID() string {
	var cpuid string
	cmd := exec.Command("wmic", "cpu", "get", "processorid")
	b, e := cmd.CombinedOutput()

	if e == nil {
		cpuid = string(b)
		cpuid = cpuid[12 : len(cpuid)-2]
		cpuid = strings.ReplaceAll(cpuid, "\n", "")
	} else {
		fmt.Printf("%v", e)
	}

	return cpuid
}

// GetBaseBoardID 获取主板的id
func GetBaseBoardID() string {
	var cpuid string
	cmd := exec.Command("wmic", "baseboard", "get", "serialnumber")
	b, e := cmd.CombinedOutput()

	if e == nil {
		cpuid = string(b)
		cpuid = cpuid[12 : len(cpuid)-2]
		cpuid = strings.ReplaceAll(cpuid, "\n", "")
	} else {
		fmt.Printf("%v", e)
	}

	return cpuid
}
