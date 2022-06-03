package sso

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestSession(t *testing.T) {
	engine := gin.Default()

	var options = Option{
		ClientID:     "5dbeab91e4904b7bae78b7e4408ceed7",
		ClientSecret: "5f1766594bf8431d87a61c71f0f859e8",
		Scope:        []string{"userinfo", "session"},
		Addr:         "http://127.0.0.1:8801",
		Retry:        2,
		Timeout:      time.Second * 2,
		Logger:       nil,
	}

	client := New(options)

	var redirect = WithRedirect(func(ctx *gin.Context, uri string, err error) {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, uri)
	})

	session := client.Session(engine, "/login", "/oauth", "/logout", redirect)
	engine.GET("/", session, func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"errno":   0,
			"message": "success",
			"data":    client.SessionUserinfo(ctx),
		})
	})

	panic(engine.Run())
}
