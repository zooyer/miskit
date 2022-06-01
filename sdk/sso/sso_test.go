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
		Addr:         "http://localhost:8801",
		Retry:        2,
		Timeout:      time.Second * 2,
		Logger:       nil,
	}

	client := New(options)

	var sessionOptions = SessionOptions{
		RedirectFunc: func(ctx *gin.Context, uri string, err error) {
			if err == nil {
				return
			}

			t.Logf(err.Error())
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"errno":   0,
				"message": "success",
				"data": gin.H{
					"redirect": uri,
				},
			})
		},
		CallbackFunc: func(ctx *gin.Context, userinfo *Userinfo, err error) {
			var (
				errno   = 0
				message = "success"
			)
			if err != nil {
				errno = 999
				message = err.Error()
				t.Logf(err.Error())
			}

			ctx.JSON(http.StatusOK, gin.H{
				"errno":   errno,
				"message": message,
				"data":    userinfo,
			})
		},
	}

	middleware, callback := client.Session(sessionOptions)
	engine.GET("/oauth/callback", callback)
	engine.Use(middleware)

	engine.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"errno":   0,
			"message": "success",
			"data":    client.SessionUserinfo(ctx),
		})
	})

	panic(engine.Run())
}
