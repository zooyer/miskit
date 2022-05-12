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

		var ctx = context.Background()

		if err = conn.Set(ctx, "name", "张三"); err != nil {
			t.Fatal(err)
		}
		value, err := conn.Get(ctx, "name")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "张三", value)

		if err = conn.SetEx(ctx, "name", "李四", time.Second); err != nil {
			t.Fatal(err)
		}
		if value, err = conn.Get(ctx, "name"); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "李四", value)

		time.Sleep(time.Second)

		if value, err = conn.Get(ctx, "name"); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "", value)

		if err = conn.Del(ctx, "name"); err != nil {
			t.Fatal(err)
		}

		if err = conn.Del(ctx, "name"); err != nil {
			t.Fatal(err)
		}
	}
}
