package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/piaohao/godis"
	"github.com/zooyer/miskit/imdb"
)

type redisConn struct {
	rds *godis.Redis
}

type redisDriver int

func (r redisDriver) parseDuration(str string) (duration time.Duration, err error) {
	if len(str) < 2 {
		return 0, fmt.Errorf("duration error")
	}

	var index = 0
	for ; index < len(str); index++ {
		if str[index] < '0' || str[index] > '9' {
			break
		}
	}

	var unit = str[index:]
	i, err := strconv.Atoi(str[:index])
	if err != nil {
		return
	}

	switch unit {
	case "ms":
		return time.Duration(i) * time.Millisecond, nil
	case "s":
		return time.Duration(i) * time.Second, nil
	case "min":
		return time.Duration(i) * time.Minute, nil
	case "h":
		return time.Duration(i) * time.Hour, nil
	}

	return 0, fmt.Errorf("unknown time unit: %s", unit)
}

func (r redisDriver) parseArgs(args string) (opts *godis.Option, err error) {
	var options = godis.Option{
		Host:              "localhost",
		Port:              6379,
		ConnectionTimeout: time.Second * 5,
		SoTimeout:         time.Second * 5,
		Password:          "",
		Db:                0,
	}

	var fields = strings.SplitN(args, "?", 2)

	var endpoint = strings.Split(fields[0], ":")
	if len(endpoint) > 0 && endpoint[0] != "" {
		options.Host = endpoint[0]
	}
	if len(endpoint) > 1 && endpoint[1] != "" {
		if options.Port, err = strconv.Atoi(endpoint[1]); err != nil {
			return nil, fmt.Errorf("port failed, %w", err)
		}
	}

	if len(fields) > 1 && fields[1] != "" {
		var params = strings.Split(fields[1], "&")
		for _, param := range params {
			if param == "" {
				continue
			}

			var kv = strings.SplitN(param, "=", 2)
			if len(kv) < 2 || kv[1] == "" {
				continue
			}

			switch kv[0] {
			case "password", "Password":
				options.Password = kv[1]
			case "db", "Db", "DB":
				if options.Db, err = strconv.Atoi(kv[1]); err != nil {
					return nil, fmt.Errorf("args error, %w", err)
				}
			case "so_timeout", "soTimeout", "SoTimeout":
				if options.SoTimeout, err = r.parseDuration(kv[1]); err != nil {
					return
				}
			case "conn_timeout", "connection_timeout", "connTimeout", "connectionTimeout", "ConnTimeout", "ConnectionTimeout":
				if options.ConnectionTimeout, err = r.parseDuration(kv[1]); err != nil {
					return
				}
			}
		}
	}

	return &options, nil
}

func (r redisDriver) Open(args string) (conn imdb.Conn, err error) {
	var c redisConn

	opts, err := r.parseArgs(args)
	if err != nil {
		return
	}

	c.rds = godis.NewRedis(opts)
	if _, err = c.rds.Ping(); err != nil {
		return
	}

	return &c, nil
}

func (c redisConn) Get(ctx context.Context, key string) (value string, err error) {
	return c.rds.Get(key)
}

func (c redisConn) Set(ctx context.Context, key, value string) (err error) {
	if _, err = c.rds.Set(key, value); err != nil {
		return
	}

	return
}

func (c redisConn) SetEx(ctx context.Context, key, value string, seconds int64) (err error) {
	if _, err = c.rds.SetEx(key, int(seconds), value); err != nil {
		return
	}

	return
}

func (c redisConn) Del(ctx context.Context, key string) (err error) {
	if _, err = c.rds.Del(key); err != nil {
		return
	}

	return
}

func (c redisConn) TTL(ctx context.Context, key string) (ttl int64, err error) {
	return c.rds.TTL(key)
}

func init() {
	imdb.Register("redis", new(redisDriver))
}
