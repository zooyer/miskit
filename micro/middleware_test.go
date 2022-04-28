package micro

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/zooyer/miskit/log"
	"github.com/zooyer/miskit/trace"
)

func TestTrace(t *testing.T) {
	engine := gin.New()
	engine.Use(Trace("test"))
	engine.GET("/trace", func(ctx *gin.Context) {
		trace := trace.Get(ctx)
		ctx.JSON(http.StatusOK, trace)
	})

	resp := get(engine, "/trace")
	data, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.NotEqual(t, "", string(data))
	t.Log(string(data))
}

func TestLogger(t *testing.T) {
	var (
		config = log.Config{
			Level: "DEBUG",
		}
		stdout    = log.NewStdoutRecorder(log.TextFormatter(true))
		logger, _ = log.New(config, nil)
	)
	logger.SetDefaultRecorder(stdout)

	engine := gin.New()
	engine.Use(Logger(logger))
	engine.GET("/logger", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, ctx.Query("action"))
	})

	resp := get(engine, "/logger?action=test")
	data, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, `"test"`, string(data))
}

func TestRecover(t *testing.T) {
	var (
		config = log.Config{
			Level: "DEBUG",
		}
		stdout    = log.NewStdoutRecorder(log.TextFormatter(true))
		logger, _ = log.New(config, nil)
	)
	logger.SetDefaultRecorder(stdout)

	engine := gin.New()
	engine.Use(Recover(logger))
	engine.GET("/recover", func(ctx *gin.Context) {
		panic("test")
	})

	resp := get(engine, "/recover")
	data, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, `"test"`, string(data))
	t.Log(string(data))
}
