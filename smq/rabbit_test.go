package smq

import (
	"fmt"
	"testing"
	"time"
)

func TestNewRabbitMQ(t *testing.T) {
	var err error
	var exchange = Exchange{
		Type:  "direct",
		Name:  "smq.name",
		Key:   "smq.key",
		Queue: "smq.queue",
	}
	client := NewRabbitMQ("amqp://user:123456@127.0.0.1:5672", exchange)
	var ch = make(chan struct{})
	if err = client.Connect(); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	go func() {
		time.Sleep(time.Second * 5)
		if err = client.Sub("", func(client Client, topic string, message []byte, options Options) (err error) {
			fmt.Println("message:", string(message))
			close(ch)
			return
		}); err != nil {
			t.Fatal(err)
		}
	}()

	if err = client.Pub("", []byte("hello,world"), TTL(6*time.Second)); err != nil {
		t.Fatal(err)
	}

	<-ch
}
