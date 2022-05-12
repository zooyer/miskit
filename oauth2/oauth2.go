package oauth2

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/zooyer/miskit/zrpc"
)

// Endpoint URL地址
type Endpoint struct {
	AuthorizeURL    string
	AccessTokenURL  string
	RefreshTokenURL string
}

// Formatter 请求参数序列化方式(JSON:Post、Form:Post、Query:Get)
type Formatter int

// Parameter 参数配置
type Parameter interface {
	// AuthCodeParams 生成请求授权码URL参数，直接拼接在地址后面
	AuthCodeParams(ctx context.Context, config Config, state string) string
	// AuthCodeTokenParams 生成授权码请求token参数
	AuthCodeTokenParams(ctx context.Context, config Config, code string) map[string]string
	// PasswordTokenParams 生成密码请求token参数
	PasswordTokenParams(ctx context.Context, config Config, username, password string) map[string]string
	// RefreshTokenParams 生成刷新token参数
	RefreshTokenParams(ctx context.Context, config Config, refreshToken string) map[string]string
	// ClientCredentialsParams 生成客户端凭证参数
	ClientCredentialsParams(ctx context.Context, config Config) map[string]string
}

// Config OAuth2服务相关配置
type Config struct {
	ClientID     string    // 应用ID
	ClientSecret string    // 应用证书
	RedirectURI  string    // 回调地址
	Scope        string    // 授权权限
	Endpoint     Endpoint  // URL地址
	Parameter    Parameter // 参数生成
	Formatter    Formatter // 序列化方式
}

// Token OAuth2 Token
type Token struct {
	AccessToken  string      // 访问Token
	TokenType    string      // Token类型
	RefreshToken string      // 刷新Token
	ExpiresIn    int64       // 过期时间(单位秒)
	Scope        string      // 授权权限
	Raw          interface{} // 自定义
}

// TokenParser Token解析
type TokenParser interface {
	Parse(data []byte) (token *Token, err error)
}

// Client OAuth2客户端
type Client struct {
	client *zrpc.Client
	config Config
}

// 序列化方式
const (
	JSONFormatter  Formatter = 1 // JSON + POST方式
	FORMFormatter  Formatter = 2 // FORM + POST方式
	QueryFormatter Formatter = 3 // Query + GET方式
)

// defaultParameter 默认标准参数生成
type defaultParameter struct{}

func (defaultParameter) AuthCodeParams(ctx context.Context, config Config, state string) string {
	var values = url.Values{
		"response_type": {"code"},
		"client_id":     {config.ClientID},
	}

	if config.RedirectURI != "" {
		values.Set("redirect_uri", config.RedirectURI)
	}

	if config.Scope != "" {
		values.Set("scope", config.Scope)
	}

	if state != "" {
		values.Set("state", state)
	}

	return values.Encode()
}

func (defaultParameter) AuthCodeTokenParams(ctx context.Context, config Config, code string) map[string]string {
	var values = map[string]string{
		"grant_type":    "authorization_code",
		"code":          code,
		"client_id":     config.ClientID,
		"client_secret": config.ClientSecret,
	}

	if config.RedirectURI != "" {
		values["redirect_uri"] = config.RedirectURI
	}

	return values
}

func (defaultParameter) PasswordTokenParams(ctx context.Context, config Config, username, password string) map[string]string {
	var values = map[string]string{
		"grant_type":    "password",
		"username":      username,
		"password":      password,
		"client_id":     config.ClientID,
		"client_secret": config.ClientSecret,
	}

	if config.Scope != "" {
		values["scope"] = config.Scope
	}

	return values
}

func (defaultParameter) RefreshTokenParams(ctx context.Context, config Config, refreshToken string) map[string]string {
	var values = map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     config.ClientID,
		"client_secret": config.ClientSecret,
	}

	return values
}

func (defaultParameter) ClientCredentialsParams(ctx context.Context, config Config) map[string]string {
	var values = map[string]string{
		"grant_type": "client_credentials",
	}

	if config.Scope != "" {
		values["scope"] = config.Scope
	}

	return values
}

// NewClient 场景OAuth2客户端
func NewClient(rpc *zrpc.Client, config Config) *Client {
	if config.Parameter == nil {
		config.Parameter = defaultParameter{}
	}

	if config.Formatter == 0 {
		config.Formatter = JSONFormatter
	}

	var client = Client{
		client: rpc,
		config: config,
	}

	return &client
}

// genState 生成随机state
func (c *Client) genState(ctx context.Context) string {
	id := uuid.NewMD5(uuid.New(), []byte(c.config.ClientID)).String()
	return strings.ReplaceAll(id, "-", "")
}

// doRequest 请求OAuth2服务
func (c *Client) doRequest(ctx context.Context, url string, params map[string]string, parser TokenParser) (*Token, error) {
	var (
		err   error
		data  []byte
		token *Token
		form  = make(map[string][]string)
	)

	for key, val := range params {
		form[key] = []string{val}
	}

	switch c.config.Formatter {
	case JSONFormatter:
		data, _, err = c.client.PostJSON(ctx, url, params, nil)
	case FORMFormatter:
		data, _, err = c.client.PostForm(ctx, url, form, nil)
	case QueryFormatter:
		data, _, err = c.client.Get(ctx, url, form, nil)
	default:
		return nil, errors.New("invalid formatter")
	}

	if err != nil {
		return nil, err
	}

	if token, err = parser.Parse(data); err != nil {
		return nil, err
	}

	return token, nil
}

// AuthCodeURL 获取授权码URL地址
func (c *Client) AuthCodeURL(ctx context.Context) (state string, url string) {
	var buf bytes.Buffer
	buf.WriteString(c.config.Endpoint.AuthorizeURL)

	if strings.Contains(c.config.Endpoint.AuthorizeURL, "?") {
		buf.WriteByte('&')
	} else {
		buf.WriteByte('?')
	}

	state = c.genState(ctx)
	params := c.config.Parameter.AuthCodeParams(ctx, c.config, state)
	buf.WriteString(params)

	return state, buf.String()
}

// AuthorizationCodeToken 授权码方式获取Token
func (c *Client) AuthorizationCodeToken(ctx context.Context, code string, parser TokenParser) (*Token, error) {
	var params = c.config.Parameter.AuthCodeTokenParams(ctx, c.config, code)
	return c.doRequest(ctx, c.config.Endpoint.AccessTokenURL, params, parser)
}

// PasswordCredentialsToken 密码方式获取Token
func (c *Client) PasswordCredentialsToken(ctx context.Context, username, password string, parser TokenParser) (*Token, error) {
	var params = c.config.Parameter.PasswordTokenParams(ctx, c.config, username, password)
	return c.doRequest(ctx, c.config.Endpoint.AccessTokenURL, params, parser)
}

// ClientCredentialsToken 客户端凭证
func (c *Client) ClientCredentialsToken(ctx context.Context, parser TokenParser) (*Token, error) {
	var params = c.config.Parameter.ClientCredentialsParams(ctx, c.config)
	return c.doRequest(ctx, c.config.Endpoint.AccessTokenURL, params, parser)
}

// RefreshToken 刷新Token
func (c *Client) RefreshToken(ctx context.Context, refreshToken string, parser TokenParser) (*Token, error) {
	var params = c.config.Parameter.RefreshTokenParams(ctx, c.config, refreshToken)
	return c.doRequest(ctx, c.config.Endpoint.RefreshTokenURL, params, parser)
}
