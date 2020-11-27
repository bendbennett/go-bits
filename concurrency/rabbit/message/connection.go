package message

import (
	"context"
	"errors"
	"fmt"
	"github.com/bendbennett/go-bits/concurrency/rabbit/config"
	"github.com/streadway/amqp"
	"log"
	"net/url"
	"time"
)

type ConnManager struct {
	conn            *amqp.Connection
	notifyConnClose chan *amqp.Error
	address         string
	retry           retry
}

type retry struct {
	interval time.Duration
	max      int
}

func NewConnManager(conf config.Connection) *ConnManager {
	address := fmt.Sprintf("%s://%s:%s@%s:%d/",
		conf.Schema,
		conf.Username,
		conf.Password,
		conf.Host,
		conf.Port,
	)

	return &ConnManager{
		address: address,
		retry: retry{
			interval: conf.Retry.Interval,
			max:      conf.Retry.Max,
		},
	}
}

func (c *ConnManager) Connect(ctx context.Context) error {
	log.Println("attempting to connect.")

	var err error

	c.conn, err = c.dial()
	if err != nil {
		return err
	}

	c.notifyConnClose = c.conn.NotifyClose(make(chan *amqp.Error))

	go c.reconnect(ctx)

	log.Println("connected.")

	return nil
}

func (c *ConnManager) reconnect(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case amqpErr := <-c.notifyConnClose:
		if amqpErr == nil {
			log.Println("connection closed explicitly.")
			return
		}
	}

	var (
		retries = 0
		err     error
	)

	for {
		if retries >= c.retry.max {
			log.Printf("abandoning reconnection, retries: %d, maxRetries: %d\n", retries, c.retry.max)
			return
		}

		log.Printf("attempting to reconnect, retry: %d\n", retries+1)

		c.conn, err = c.dial()
		if err != nil {
			log.Println("failed to reconnect, retrying...")

			select {
			case <-ctx.Done():
				return
			case <-time.After(c.retry.interval):
				retries++
				continue
			}
		}

		c.notifyConnClose = c.conn.NotifyClose(make(chan *amqp.Error))

		log.Println("reconnected")

		c.reconnect(ctx)
	}
}

func (c *ConnManager) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *ConnManager) Channel() (*amqp.Channel, error) {
	if c.conn == nil {
		return nil, errors.New("no connection")
	}
	return c.conn.Channel()
}

func (c *ConnManager) dial() (*amqp.Connection, error) {
	addressURL, err := url.Parse(c.address)
	if err != nil {
		return nil, err
	}

	if addressURL.Scheme == "amqps" {
		panic("implement me")
	}

	return amqp.Dial(addressURL.String())
}
