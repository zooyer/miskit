package oauth2

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/zooyer/miskit/log"
	"github.com/zooyer/miskit/zrpc"
)

type GithubError struct {
	ErrorName        string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

func (e GithubError) Error() string {
	return e.ErrorDescription
}

type GithubToken struct {
	*GithubError
	AccessToken string          `json:"access_token"`
	TokenType   string          `json:"token_type"`
	Scope       string          `json:"scope,omitempty"`
	Raw         json.RawMessage `json:"raw,omitempty"`
}

func (t *GithubToken) Parse(data []byte) (token *Token, err error) {
	//values, err := url.ParseQuery(string(data))
	//if err != nil {
	//	return
	//}
	//
	//t.AccessToken = values.Get("access_token")
	//t.TokenType = values.Get("token_type")
	//t.Scope = values.Get("scope")
	//t.Raw = data

	t.Raw = data
	if err = json.Unmarshal(data, t); err != nil {
		return
	}

	if t.ErrorName != "" {
		return nil, t.GithubError
	}

	token = new(Token)
	token.AccessToken = t.AccessToken
	token.TokenType = t.TokenType
	token.Scope = t.Scope
	token.Raw = t

	return
}

type wechat struct{}

func (wechat) AuthCodeParams(ctx context.Context, config Config, state string) string {
	var values = url.Values{
		"appid":         {config.ClientID},
		"response_type": {"code"},
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

	return values.Encode() + "#wechat_redirect"
}

func (wechat) AuthCodeTokenParams(ctx context.Context, config Config, code string) map[string]string {
	var values = map[string]string{
		"grant_type": "authorization_code",
		"code":       code,
		"appid":      config.ClientID,
		"secret":     config.ClientSecret,
	}

	return values
}

func (wechat) PasswordTokenParams(ctx context.Context, config Config, username, password string) map[string]string {
	panic("implement")
}

func (wechat) RefreshTokenParams(ctx context.Context, config Config, refreshToken string) map[string]string {
	var values = map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"appid":         config.ClientID,
	}

	return values
}

func (wechat) ClientCredentialsParams(ctx context.Context, config Config) map[string]string {
	panic("implement")
}

type WechatError struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

func (e WechatError) Error() string {
	return e.Errmsg
}

type WechatToken struct {
	*WechatError
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	Openid       string `json:"openid"`
	Scope        string `json:"scope"`
}

func (wt *WechatToken) Parse(data []byte) (token *Token, err error) {
	if err = json.Unmarshal(data, wt); err != nil {
		return
	}

	if wt.WechatError != nil {
		return nil, wt.WechatError
	}

	token = &Token{
		AccessToken:  wt.AccessToken,
		TokenType:    "",
		RefreshToken: wt.RefreshToken,
		ExpiresIn:    wt.ExpiresIn,
		Scope:        wt.Scope,
		Raw:          data,
	}

	return
}

func TestGithub(t *testing.T) {
	var step = 1
	var config = Config{
		ClientID:     "42c54212574100f363a25b3",
		ClientSecret: "24755362f634b63e30c09412b4d99e20bf6d2593",
		Endpoint: Endpoint{
			AuthorizeURL:    "https://github.com/login/oauth/authorize",
			AccessTokenURL:  "https://github.com/login/oauth/access_token",
			RefreshTokenURL: "https://github.com/login/oauth/access_token",
		},
		RedirectURI: "http://i.genkitol.com:8000/oauth/callback",
		Scope:       "user",
	}

	var (
		ctx = context.Background()
		cfg = log.Config{
			Output: "stdout",
			Level:  "DEBUG",
		}
		logger, _ = log.New(cfg, log.TextFormatter(true))
		rpc       = zrpc.New("test", 2, time.Second*5, logger, func(ctx context.Context, req *zrpc.Request) {
			req.Header.Set("Accept", "application/json")
		})
	)

	var client = NewClient(rpc, config)

	if step == 1 {
		t.Log(client.AuthCodeURL(ctx))
	}

	if step == 2 {
		var code = "7fd9f85398f7fe6955fe"

		var tk GithubToken

		token, err := client.AuthorizationCodeToken(ctx, code, &tk)
		if err != nil {
			t.Logf("%#v", err)
			t.Fatal(err)
		}

		var data []byte

		if data, err = json.Marshal(token); err != nil {
			t.Fatal(err)
		}
		t.Log(string(data))

		if data, err = json.Marshal(tk); err != nil {
			t.Fatal(err)
		}
		t.Log(string(data))
	}
}

func TestWechat(t *testing.T) {
	var step = 3
	var config = Config{
		ClientID:     "wxd14f8452d626921567",
		ClientSecret: "0a44a854115b0e635124afd24183e7",
		Endpoint: Endpoint{
			AuthorizeURL:    "https://open.weixin.qq.com/connect/oauth2/authorize",
			AccessTokenURL:  "https://api.weixin.qq.com/sns/oauth2/access_token",
			RefreshTokenURL: "https://api.weixin.qq.com/sns/oauth2/refresh_token",
			// 验证token https://api.weixin.qq.com/sns/auth
		},
		RedirectURI: "http://i.genkitol.com/oauth/callback",
		Scope:       "snsapi_userinfo",
		Formatter:   QueryFormatter,
		Parameter:   wechat{},
	}

	var (
		ctx = context.Background()
		cfg = log.Config{
			Output: "stdout",
			Level:  "DEBUG",
		}
		logger, _ = log.New(cfg, log.TextFormatter(true))
		rpc       = zrpc.New("test", 2, time.Second*5, logger, func(ctx context.Context, req *zrpc.Request) {
			req.Header.Set("Accept", "application/json")
		})
	)

	var client = NewClient(rpc, config)
	if step == 1 {
		t.Log(client.AuthCodeURL(ctx))
	}

	if step == 2 {
		var code = "nzC9D1a051a8f80q6uFrjFa1xo4arjFq"

		var tk WechatToken

		token, err := client.AuthorizationCodeToken(ctx, code, &tk)
		if err != nil {
			t.Logf("%#v", err)
			t.Fatal(err)
		}

		var data []byte

		if data, err = json.Marshal(token); err != nil {
			t.Fatal(err)
		}
		t.Log(string(data))

		if data, err = json.Marshal(tk); err != nil {
			t.Fatal(err)
		}
		t.Log(string(data))
	}

	if step == 3 {
		const refreshToken = "56fpXAlVQUbrZcleT0IjnJcHtbcXPLQ_E71lsfHSfdtupKSFJS28HH_TnFDhda_lV8T6e6wPf6AAtTgeA"

		var tk WechatToken
		token, err := client.RefreshToken(ctx, refreshToken, &tk)
		if err != nil {
			t.Logf("%#v", err)
			t.Fatal(err)
		}

		var data []byte

		if data, err = json.Marshal(token); err != nil {
			t.Fatal(err)
		}
		t.Log(string(data))

		if data, err = json.Marshal(tk); err != nil {
			t.Fatal(err)
		}
		t.Log(string(data))
	}
}
