package micro

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	_ "github.com/zooyer/miskit/imdb/driver/buntdb"
	_ "github.com/zooyer/miskit/imdb/driver/redis"
)

func TestNewStore(t *testing.T) {
	store, err := NewStore("redis", "", "", []byte("test"))
	if err != nil {
		t.Fatal(err)
	}

	session := sessions.Sessions("miskit-test", store)

	engine := gin.Default()
	engine.Use(session)

	engine.GET("/set", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		session.Set(ctx.Query("key"), ctx.Query("value"))
		if err := session.Save(); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			ctx.Abort()
			return
		}

		ctx.String(http.StatusOK, "ok")
	})

	engine.GET("/get", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		value := session.Get(ctx.Query("key"))
		ctx.String(http.StatusOK, "%s", value)
	})

	const (
		key   = "name"
		value = "张三"
	)

	resp := get(engine, fmt.Sprintf("/set?key=%s&value=%s", key, value))
	data, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "ok", string(data))

	cookie := resp.Header().Get("Set-Cookie")
	resp = get(engine, fmt.Sprintf("/get?key=%s", key), http.Header{
		"Cookie": {cookie},
	})
	data, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, value, string(data))
}
