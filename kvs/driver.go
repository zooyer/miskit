package kvs

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var (
	drivers = make(map[string]Driver)
	mutex   sync.Mutex
)

type Conn interface {
	Get(ctx context.Context, key string) (value string, err error)
	Set(ctx context.Context, key, value string) (err error)
	SetEx(ctx context.Context, key, value string, ttl time.Duration) (err error)
	Del(ctx context.Context, key string) (err error)
}

type Driver interface {
	Open(args string) (conn Conn, err error)
}

func Register(name string, driver Driver) {
	mutex.Lock()
	defer mutex.Unlock()

	if driver == nil {
		panic("kvs: Register driver is nil")
	}

	if _, dup := drivers[name]; dup {
		panic("kvs: Register called twice for driver " + name)
	}

	drivers[name] = driver
}

func Open(dialect string, args string) (conn Conn, err error) {
	driver, exists := drivers[dialect]
	if !exists || driver == nil {
		return nil, fmt.Errorf("kvs: unknown driver %q (forgotten import?)", dialect)
	}

	return driver.Open(args)
}
