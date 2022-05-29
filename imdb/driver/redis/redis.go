package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/zooyer/miskit/imdb"
)

type redisConn struct {
	rds *redis.Client
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

func (r redisDriver) parseArgs(args string) (opts *redis.Options, err error) {
	var options = redis.Options{
		Addr:         "localhost",
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	}

	var fields = strings.SplitN(args, "?", 2)

	var endpoint = strings.Split(fields[0], "/")
	if len(endpoint) > 0 && endpoint[0] != "" {
		options.Addr = endpoint[0]
		if len(endpoint) > 1 && endpoint[1] != "" {
			if options.DB, err = strconv.Atoi(endpoint[1]); err != nil {
				return
			}
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
			case "dial_timeout", "dialTimeout", "DialTimeout":
				if options.DialTimeout, err = r.parseDuration(kv[1]); err != nil {
					return
				}
			case "read_timeout", "readTimeout", "ReadTimeout":
				if options.ReadTimeout, err = r.parseDuration(kv[1]); err != nil {
					return
				}
			case "write_timeout", "writeTimeout", "WriteTimeout":
				if options.WriteTimeout, err = r.parseDuration(kv[1]); err != nil {
					return
				}
			case "pool_size", "poolSize", "PoolSize":
				if options.PoolSize, err = strconv.Atoi(kv[1]); err != nil {
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

	c.rds = redis.NewClient(opts)

	if _, err = c.rds.Ping().Result(); err != nil {
		return
	}

	return &c, nil
}

func (c redisConn) Get(ctx context.Context, key string) (value string, err error) {
	if value, err = c.rds.Get(key).Result(); err != nil {
		if err == redis.Nil {
			err = nil
		}
	}

	return
}

func (c redisConn) Set(ctx context.Context, key, value string) (err error) {
	if _, err = c.rds.Set(key, value, 0).Result(); err != nil {
		return
	}

	return
}

func (c redisConn) SetEx(ctx context.Context, key, value string, seconds int64) (err error) {
	if _, err = c.rds.Set(key, value, time.Second*time.Duration(seconds)).Result(); err != nil {
		return
	}

	return
}

func (c redisConn) Del(ctx context.Context, key string) (err error) {
	if _, err = c.rds.Del(key).Result(); err != nil {
		return
	}

	return
}

func (c redisConn) TTL(ctx context.Context, key string) (seconds int64, err error) {
	ttl, err := c.rds.TTL(key).Result()
	if err != nil {
		return
	}

	return int64(ttl.Seconds()), nil
}

func init() {
	imdb.Register("redis", new(redisDriver))
}
