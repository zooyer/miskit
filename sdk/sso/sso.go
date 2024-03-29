package sso

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zooyer/miskit/log"
	"github.com/zooyer/miskit/zrpc"
)

type Option struct {
	ClientID     string
	ClientSecret string
	Scope        []string
	Addr         string
	Retry        int
	Timeout      time.Duration
	Logger       *log.Logger
}

type Client struct {
	option Option
	client *zrpc.Client
}

type Pages Client

type cookie struct {
	Cookie   string `json:"cookie"`
	MaxAge   int    `json:"max_age"`
	Path     string `json:"path"`
	Domain   string `json:"domain"`
	Secure   bool   `json:"secure"`
	HttpOnly bool   `json:"http_only"`
}

type sessionResp struct {
	Cookie   *cookie   `json:"cookie"`
	Userinfo *Userinfo `json:"userinfo"`
}

type Token struct {
	AccessToken  string `json:"access_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

type Userinfo struct {
	UserID        int64  `json:"user_id"`
	Username      string `json:"username,omitempty"`
	Nickname      string `json:"nickname,omitempty"`
	Realname      string `json:"realname,omitempty"`
	UserPhone     string `json:"user_phone,omitempty"`
	UserEmail     string `json:"user_email,omitempty"`
	UserGender    int    `json:"user_gender,omitempty"`
	UserAddress   string `json:"user_address,omitempty"`
	UserAvatar    string `json:"user_avatar,omitempty"`
	UserAccess    string `json:"user_access,omitempty"`
	UserSource    string `json:"user_source,omitempty"`
	UserStatus    int    `json:"user_status,omitempty"`
	UserExpiredAt int64  `json:"user_expired_at,omitempty"`
}

type Router interface {
	gin.IRouter
	BasePath() string
}

func New(option Option) *Client {
	return &Client{
		option: option,
		client: zrpc.New("sso", option.Retry, option.Timeout, option.Logger),
	}
}

func (c *Client) AuthorizeCodeURL(ctx context.Context, redirectURI string) string {
	var params = url.Values{
		"response_type": {"code"},
		"client_id":     {c.option.ClientID},
	}

	if redirectURI != "" {
		params.Set("redirect_uri", redirectURI)
	}

	if len(c.option.Scope) > 0 {
		params.Set("scope", strings.Join(c.option.Scope, " "))
	}

	// TODO 后续考虑增加state、code_challenge、code_challenge_method
	return fmt.Sprintf("%v/sso/authorize?%s", c.option.Addr, params.Encode())
}

func (c *Client) Token(ctx context.Context, code string) (_ *Token, err error) {
	var req = map[string]string{
		"grant_type":    "authorization_code",
		"code":          code,
		"client_id":     c.option.ClientID,
		"client_secret": c.option.ClientSecret,
		// TODO 后续增加code_verifier，对code_challenge做校验
	}

	var (
		uri   = fmt.Sprintf("%v/sso/api/v1/oauth/token", c.option.Addr)
		token Token
	)

	if _, _, err = c.client.PostJSON(ctx, uri, req, &token); err != nil {
		return
	}

	return &token, nil
}

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (_ *Token, err error) {
	var req = map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     c.option.ClientID,
		"client_secret": c.option.ClientSecret,
	}

	var (
		uri   = fmt.Sprintf("%v/sso/api/v1/oauth/token", c.option.Addr)
		token Token
	)

	if _, _, err = c.client.PostJSON(ctx, uri, req, &token); err != nil {
		return
	}

	return &token, nil
}

func (c *Client) Verify(ctx context.Context, accessToken string) (err error) {
	var (
		uri = fmt.Sprintf("%v/sso/api/v1/oauth/verify", c.option.Addr)
		req = map[string]interface{}{
			"client_id":    c.option.ClientID,
			"access_token": accessToken,
		}
		resp interface{}
	)

	if _, _, err = c.client.PostJSON(ctx, uri, req, &resp); err != nil {
		return
	}

	return nil
}

func (c *Client) Userinfo(ctx *gin.Context, accessToken string) (userinfo *Userinfo, err error) {
	var (
		uri = fmt.Sprintf("%v/sso/api/v1/oauth/userinfo", c.option.Addr)
		req = map[string]interface{}{
			"client_id":    c.option.ClientID,
			"access_token": accessToken,
		}
		resp Userinfo
	)

	if _, _, err = c.client.PostJSON(ctx, uri, req, &resp); err != nil {
		return
	}

	return c.userinfo(ctx, &resp), nil
}

func (c *Client) cookieName() string {
	return fmt.Sprintf("sso-cookie-%v", c.option.ClientID)
}

func (c *Client) contextKey() string {
	return fmt.Sprintf("sso-session-%v", c.option.ClientID)
}

// setCookie 设置cookie
func (c *Client) setCookie(ctx *gin.Context, cookie *cookie) {
	if cookie == nil {
		return
	}

	ctx.SetCookie(
		c.cookieName(),
		cookie.Cookie,
		cookie.MaxAge,
		cookie.Path,
		cookie.Domain,
		cookie.Secure,
		cookie.HttpOnly,
	)
}

// getSession 获取session
func (c *Client) getSession(ctx *gin.Context, cookie string) (session *sessionResp, err error) {
	var (
		uri = fmt.Sprintf("%v/sso/api/v1/oauth/session/get", c.option.Addr)
		req = map[string]interface{}{
			"client_id": c.option.ClientID,
			"cookie":    cookie,
		}
		resp sessionResp
	)

	if _, _, err = c.client.PostJSON(ctx, uri, req, &resp); err != nil {
		return
	}

	return &resp, nil
}

// newSession 创建session
func (c *Client) newSession(ctx *gin.Context, code string) (session *sessionResp, err error) {
	token, err := c.Token(ctx, code)
	if err != nil {
		return
	}

	var (
		uri = fmt.Sprintf("%v/sso/api/v1/oauth/session/new", c.option.Addr)
		req = map[string]interface{}{
			"client_id":    c.option.ClientID,
			"access_token": token.AccessToken,
		}
		resp sessionResp
	)

	if _, _, err = c.client.PostJSON(ctx, uri, req, &resp); err != nil {
		return
	}

	return &resp, nil
}

// delSession 删除session
func (c *Client) delSession(ctx *gin.Context, cookie string) (session *sessionResp, err error) {
	var (
		uri = fmt.Sprintf("%v/sso/api/v1/oauth/session/del", c.option.Addr)
		req = map[string]interface{}{
			"client_id": c.option.ClientID,
			"cookie":    cookie,
		}
		resp sessionResp
	)

	if _, _, err = c.client.PostJSON(ctx, uri, req, &resp); err != nil {
		return
	}

	return &resp, nil
}

type sessionOptions struct {
	// 用户未登录重定向到登录页，默认未登录则会302重定向到登录页
	RedirectFunc func(ctx *gin.Context, uri string, err error)
	// 用户授权登录后回调，默认失败会返回403状态码
	CallbackFunc func(ctx *gin.Context, userinfo *Userinfo, err error)
	// 用户注销登录，默认失败会返回
	LogoutFunc func(ctx *gin.Context, err error)
}

type SessionOption func(options *sessionOptions)

func WithRedirect(redirect func(ctx *gin.Context, uri string, err error)) SessionOption {
	return func(options *sessionOptions) {
		options.RedirectFunc = redirect
	}
}

func WithCallback(callback func(ctx *gin.Context, userinfo *Userinfo, err error)) SessionOption {
	return func(options *sessionOptions) {
		options.CallbackFunc = callback
	}
}

func WithLogout(logout func(ctx *gin.Context, err error)) SessionOption {
	return func(options *sessionOptions) {
		options.LogoutFunc = logout
	}
}

func (c *Client) userinfo(ctx *gin.Context, userinfo *Userinfo) *Userinfo {
	if userinfo == nil {
		return nil
	}

	if userinfo.UserAvatar != "" {
		userinfo.UserAvatar = fmt.Sprintf("%s%s", c.option.Addr, userinfo.UserAvatar)
	}

	return userinfo
}

func (c *Client) oauth(options sessionOptions) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			err error
			req struct {
				Code string `form:"code" json:"code" binding:"required"`
			}
			resp     *sessionResp
			userinfo *Userinfo
		)

		defer func() {
			if options.CallbackFunc != nil {
				options.CallbackFunc(ctx, userinfo, err)
			} else {
				if err != nil {
					ctx.AbortWithStatus(http.StatusForbidden)
				}
			}
		}()

		if err = ctx.Bind(&req); err != nil {
			return
		}

		if resp, err = c.newSession(ctx, req.Code); err != nil {
			return
		}

		if userinfo = resp.Userinfo; userinfo != nil {
			userinfo = c.userinfo(ctx, userinfo)
		}

		if resp.Cookie != nil {
			c.setCookie(ctx, resp.Cookie)
		}
	}
}

func (c *Client) middleware(loginPath string, options sessionOptions) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			err    error
			cookie string
		)

		defer func() {
			if err != nil {
				if options.RedirectFunc != nil {
					options.RedirectFunc(ctx, loginPath, err)
				} else {
					ctx.Redirect(http.StatusFound, loginPath)
				}
				ctx.Abort()
			}
		}()

		if cookie, err = ctx.Cookie(c.cookieName()); err != nil {
			return
		}

		session, err := c.getSession(ctx, cookie)
		if err != nil {
			return
		}

		ctx.Set(c.contextKey(), c.userinfo(ctx, session.Userinfo))

		if session.Cookie != nil {
			c.setCookie(ctx, session.Cookie)
		}
	}
}

func (c *Client) login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var authCodeURL = c.AuthorizeCodeURL(ctx, ctx.Query("redirect_uri"))
		uri, err := url.Parse(authCodeURL)
		if err != nil {
			ctx.Redirect(http.StatusFound, authCodeURL)
			return
		}

		var query = uri.Query()
		for key, values := range ctx.Request.URL.Query() {
			for _, value := range values {
				query.Set(key, value)
			}
		}

		uri.RawQuery = query.Encode()
		uri.Fragment = ctx.Request.URL.Fragment

		ctx.Redirect(http.StatusFound, uri.String())
	}
}

func (c *Client) logout(options sessionOptions) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			err    error
			cookie string
		)

		defer func() {
			if options.LogoutFunc != nil {
				options.LogoutFunc(ctx, err)
			} else {
				if err != nil {
					ctx.AbortWithStatus(http.StatusForbidden)
				}
			}
		}()

		if cookie, err = ctx.Cookie(c.cookieName()); err != nil {
			return
		}

		session, err := c.delSession(ctx, cookie)
		if err != nil {
			return
		}

		if session.Cookie != nil {
			c.setCookie(ctx, session.Cookie)
		}
	}
}

func (c *Client) Session(router Router, loginPath, oauthPath, logoutPath string, options ...SessionOption) (middleware gin.HandlerFunc) {
	var opt sessionOptions
	for _, fn := range options {
		fn(&opt)
	}

	middleware = c.middleware(path.Join(router.BasePath(), loginPath), opt)

	router.GET(loginPath, c.login())
	router.HEAD(loginPath, c.login())
	router.GET(oauthPath, c.oauth(opt))
	router.POST(oauthPath, c.oauth(opt))
	router.GET(logoutPath, middleware, c.logout(opt))
	router.POST(logoutPath, middleware, c.logout(opt))

	return middleware
}

func (c *Client) SessionUserinfo(ctx context.Context) *Userinfo {
	if userinfo, ok := ctx.Value(c.contextKey()).(*Userinfo); ok {
		return userinfo
	}

	return nil
}

func (c *Client) Pages() *Pages {
	return (*Pages)(c)
}

func (p *Pages) Login() string {
	return fmt.Sprintf("%v/login", p.option.Addr)
}

func (p *Pages) Home() string {
	return fmt.Sprintf("%v", p.option.Addr)
}

func (p *Pages) Dashboard() string {
	return fmt.Sprintf("%v/dashboard", p.option.Addr)
}

func (p *Pages) Profile() string {
	return fmt.Sprintf("%v/profile", p.option.Addr)
}
