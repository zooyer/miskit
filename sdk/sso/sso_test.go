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
		ClientID:     "c51a92a640a04d26adcdb0f26c517487",
		ClientSecret: "e0184070867e409e9b124b834475609d",
		Addr:         "http://192.168.1.10:8881",
		Retry:        2,
		Timeout:      time.Second * 2,
		Logger:       nil,
	}

	client := New(options)

	var redirect = WithRedirect(func(ctx *gin.Context, uri string, err error) {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, uri)
	})

	session := client.Session(engine, "/login", "/oauth", redirect)
	engine.GET("/", session, func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"errno":   0,
			"message": "success",
			"data":    client.SessionUserinfo(ctx),
		})
	})

	panic(engine.Run())
}
