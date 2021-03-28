package smq

import (
	"time"
)

type Options struct {
	Ack bool
	TTL time.Duration
}

type Option func(ops *Options)

type Subscriber func(client Client, topic string, message []byte, options Options) (err error)

type Client interface {
	Connect() error
	Pub(topic string, message []byte, option ...Option) error
	Sub(topic string, subscriber Subscriber) error
	Close() error
}

func TTL(ttl time.Duration) Option {
	return func(ops *Options) {
		ops.TTL = ttl
	}
}

func Ack(ack bool) Option {
	return func(ops *Options) {
		ops.Ack = ack
	}
}

func options(option ...Option) Options {
	var options Options
	for _, opt := range option {
		opt(&options)
	}
	return options
}
