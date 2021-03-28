package conf

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Conf struct {
	mutex    sync.Mutex
	viper    *viper.Viper
	ticker   *time.Ticker
	interval time.Duration
	addr     map[string]interface{}
}

func New(interval time.Duration) *Conf {
	var conf Conf
	conf.interval = interval
	conf.viper = viper.New()
	conf.addr = make(map[string]interface{})

	if interval > 0 {
		conf.ticker = time.NewTicker(interval)
		go conf.reload()
	}

	return &conf
}

func (c *Conf) toValue() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for key, addr := range c.addr {
		_ = c.viper.UnmarshalKey(key, addr)
	}
}

func (c *Conf) reload() {
	for range c.ticker.C {
		c.toValue()
	}
}

func (c *Conf) key(key ...interface{}) string {
	var keys = make([]string, 0, len(key))
	for _, key := range key {
		keys = append(keys, fmt.Sprint(key))
	}
	return strings.Join(keys, ".")
}

func (c *Conf) Stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.ticker != nil {
		c.ticker.Stop()
	}
}

func (c *Conf) Bind(key string, v interface{}) (err error) {
	if err := c.viper.UnmarshalKey(key, v); err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.interval > 0 {
		c.addr[key] = v
	}

	return
}

func (c *Conf) Init(filename string) (err error) {
	c.viper.SetConfigFile(filename)
	if c.interval > 0 {
		c.viper.WatchConfig()
	}
	if err = c.viper.ReadInConfig(); err != nil {
		return
	}
	return
}
