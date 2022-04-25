package driver

import (
	"context"
	"testing"
	"time"

	"github.com/zooyer/miskit/kvs"
	"github.com/zooyer/miskit/kvs/driver/buntdb"
	"github.com/zooyer/miskit/kvs/driver/redis"
)

var (
	_ redis.RedisDriver
	_ buntdb.BuntdbDriver
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
		conn, err := kvs.Open(test.dialect, test.args)
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

		if value != "张三" {
			t.Fatal("value:", value, "!= \"张三\"")
		}

		if err = conn.SetEx(ctx, "name", "李四", time.Second); err != nil {
			t.Fatal(err)
		}

		if value, err = conn.Get(ctx, "name"); err != nil {
			t.Fatal(err)
		}
		if value != "李四" {
			t.Fatal("value:", value, "!= \"李四\"")
		}

		time.Sleep(time.Second)

		if value, err = conn.Get(ctx, "name"); err != nil {
			t.Fatal(err)
		}

		if value != "" {
			t.Fatal("value:", value, "!= \"\"")
		}
	}
}
