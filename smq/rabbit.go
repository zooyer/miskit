package smq

import (
	"fmt"
	"github.com/streadway/amqp"
	"sync"
)

type Exchange struct {
	Type  string // 交换机类型
	Name  string // 交换机名称
	Key   string // key值
	Queue string // 队列名称
}

type rabbit struct {
	Exchange
	mutex      sync.RWMutex
	broker     string
	channel    *amqp.Channel
	connection *amqp.Connection
}

func NewRabbitMQ(broker string, exchange Exchange) *rabbit {
	return &rabbit{
		broker:   broker,
		Exchange: exchange,
	}
}

func (r *rabbit) Connect() (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.connection == nil {
		if r.connection, err = amqp.Dial(r.broker); err != nil {
			return
		}
	}

	if r.channel == nil {
		if r.channel, err = r.connection.Channel(); err != nil {
			return
		}
	}

	return
}

func (r *rabbit) Pub(topic string, message []byte, option ...Option) (err error) {
	options := options(option...)

	// 用于检查队列是否存在,已经存在不需要重复声明
	if _, err = r.channel.QueueDeclarePassive(r.Queue, true, false, false, true, nil); err != nil {
		// 队列不存在,声明队列
		// name:队列名称;durable:是否持久化,队列存盘,true服务重启后信息不会丢失,影响性能;autoDelete:是否自动删除;noWait:是否非阻塞,
		// true为是,不等待RMQ返回信息;args:参数,传nil即可;exclusive:是否设置排他
		if _, err = r.channel.QueueDeclare(r.Queue, true, false, false, true, nil); err != nil {
			return
		}
	}

	// 队列绑定
	if err = r.channel.QueueBind(r.Queue, r.Key, r.Name, true, nil); err != nil {
		return
	}

	// 用于检查交换机是否存在,已经存在不需要重复声明
	if err = r.channel.ExchangeDeclarePassive(r.Name, r.Type, true, false, false, true, nil); err != nil {
		// 注册交换机
		// name:交换机名称,kind:交换机类型,durable:是否持久化,队列存盘,true服务重启后信息不会丢失,影响性能;autoDelete:是否自动删除;
		// noWait:是否非阻塞, true为是,不等待RMQ返回信息;args:参数,传nil即可; internal:是否为内部
		if err = r.channel.ExchangeDeclare(r.Name, r.Type, true, false, false, true, nil); err != nil {
			return
		}
	}

	// 发送任务消息
	var publish = amqp.Publishing{
		ContentType: "text/plain",
		Body:        message,
		Expiration:  fmt.Sprint(options.TTL.Milliseconds()),
	}
	if err = r.channel.Publish(r.Name, r.Key, false, false, publish); err != nil {
		return
	}

	return
}

func (r *rabbit) Sub(topic string, subscriber Subscriber) (err error) {
	// 用于检查队列是否存在,已经存在不需要重复声明
	if _, err = r.channel.QueueDeclarePassive(r.Queue, true, false, false, true, nil); err != nil {
		// 队列不存在,声明队列
		// name:队列名称;durable:是否持久化,队列存盘,true服务重启后信息不会丢失,影响性能;autoDelete:是否自动删除;noWait:是否非阻塞,
		// true为是,不等待RMQ返回信息;args:参数,传nil即可;exclusive:是否设置排他
		if _, err = r.channel.QueueDeclare(r.Queue, true, false, false, true, nil); err != nil {
			return
		}
	}

	// 队列绑定
	if err = r.channel.QueueBind(r.Queue, r.Key, r.Name, true, nil); err != nil {
		return
	}

	// 获取消费通道,确保rabbitMQ一个一个发送消息
	if err = r.channel.Qos(1, 0, true); err != nil {
		return
	}
	consume, err := r.channel.Consume(r.Queue, "", false, false, false, false, nil)
	if err != nil {
		return
	}

	go func() {
		for msg := range consume {
			var multiple bool
			if err = subscriber(r, topic, msg.Body, Options{}); err != nil {
				multiple = true
				// TODO callback error
			}
			if err = msg.Ack(multiple); err != nil {
				// TODO ack error
			}
		}
	}()

	return
}

func (r *rabbit) Close() (err error) {
	if err = r.channel.Close(); err != nil {
		return
	}

	if err = r.connection.Close(); err != nil {
		return
	}

	return
}
