package imdb

import (
	"context"
	"fmt"
	"sync"
)

var (
	drivers = make(map[string]Driver)
	mutex   sync.Mutex
)

type Conn interface {
	Get(ctx context.Context, key string) (value string, err error)
	Set(ctx context.Context, key, value string) (err error)
	SetEx(ctx context.Context, key, value string, seconds int64) (err error)
	Del(ctx context.Context, key string) (err error)
	TTL(ctx context.Context, key string) (seconds int64, err error)
}

type Driver interface {
	Open(args string) (conn Conn, err error)
}

func Register(name string, driver Driver) {
	mutex.Lock()
	defer mutex.Unlock()

	if driver == nil {
		panic("imdb: Register driver is nil")
	}

	if _, dup := drivers[name]; dup {
		panic("imdb: Register called twice for driver " + name)
	}

	drivers[name] = driver
}

func Open(dialect string, args string) (conn Conn, err error) {
	driver, exists := drivers[dialect]
	if !exists || driver == nil {
		return nil, fmt.Errorf("imdb: unknown driver %q (forgotten import?)", dialect)
	}

	return driver.Open(args)
}
