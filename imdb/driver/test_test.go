package driver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zooyer/miskit/imdb"
	_ "github.com/zooyer/miskit/imdb/driver/buntdb"
	_ "github.com/zooyer/miskit/imdb/driver/redis"
)

func TestOpen(t *testing.T) {
	var tests = []struct {
		dialect string
		args    string
	}{
		{
			dialect: "buntdb",
			args:    ":memory:",
		},
		{
			dialect: "redis",
			args:    "localhost:6379",
		},
	}

	for _, test := range tests {
		conn, err := imdb.Open(test.dialect, test.args)
		if err != nil {
			t.Fatal(err)
		}

		const key = "name"
		var ctx = context.Background()

		if err = conn.Set(ctx, key, "张三"); err != nil {
			t.Fatal(err)
		}
		value, err := conn.Get(ctx, key)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "张三", value)

		if err = conn.SetEx(ctx, key, "李四", 1); err != nil {
			t.Fatal(err)
		}
		if value, err = conn.Get(ctx, key); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "李四", value)

		time.Sleep(time.Second)

		if value, err = conn.Get(ctx, key); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "", value)

		if err = conn.Del(ctx, key); err != nil {
			t.Fatal(err)
		}

		if err = conn.Del(ctx, key); err != nil {
			t.Fatal(err)
		}

		seconds, err := conn.TTL(ctx, key)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int64(-2), seconds)

		if err = conn.SetEx(ctx, key, "ttl", 1); err != nil {
			t.Fatal(err)
		}

		if seconds, err = conn.TTL(ctx, key); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int64(1), seconds)
	}
}
