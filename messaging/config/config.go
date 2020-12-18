package config

import (
	"github.com/streadway/amqp"
	"time"
)

type Config struct {
	Messaging Messaging
}

type Messaging struct {
	Connection Connection
	Exchange   Exchange
	Queue      Queue
	QueueBind  QueueBind
	Publish    Publish
}

type Connection struct {
	Schema   string
	Username string
	Password string
	Host     string
	Port     int
	Retry    Retry
}

type Retry struct {
	Max      int
	Interval time.Duration
}

type Exchange struct {
	Name       string
	Kind       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       amqp.Table
}

type Queue struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

type QueueBind struct {
	Name     string
	Key      string
	Exchange string
	NoWait   bool
	Args     amqp.Table
}

type Publish struct {
	Timeout time.Duration
}

func New() Config {
	return Config{
		Messaging: Messaging{
			Connection: Connection{
				Schema:   "amqp",
				Username: "guest",
				Password: "guest",
				Host:     "localhost",
				Port:     5672,
				Retry: Retry{
					Max:      100,
					Interval: 1 * time.Second,
				},
			},
			Exchange: Exchange{
				Name:       "my_exchange",
				Kind:       "direct",
				Durable:    true,
				AutoDelete: false,
				Internal:   false,
				NoWait:     false,
				Args:       nil,
			},
			Queue: Queue{
				Name:       "my_queue",
				Durable:    true,
				AutoDelete: false,
				Exclusive:  false,
				NoWait:     false,
				Args:       nil,
			},
			QueueBind: QueueBind{
				Name:     "my_queue",
				Key:      "my_routing_key",
				Exchange: "my_exchange",
				NoWait:   false,
				Args:     nil,
			},
			Publish: Publish{
				Timeout: 1 * time.Second,
			},
		},
	}
}
