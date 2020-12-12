package message

import (
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

type Publisher struct {
	getChannel GetChannel
	exchange   string
	routingKey string
	timeout    time.Duration
}

func NewPublisher(getChannel GetChannel, exchange, routingKey string, timeout time.Duration) *Publisher {
	return &Publisher{
		getChannel: getChannel,
		exchange:   exchange,
		routingKey: routingKey,
		timeout:    timeout,
	}
}

func (p *Publisher) Publish(body string) error {
	channel, err := p.getChannel()
	if err != nil {
		return fmt.Errorf("cannot get channel: %w", err)
	}
	defer channel.Close()

	err = channel.Confirm(false)
	if err != nil {
		return fmt.Errorf("cannot set confirm: %w", err)
	}

	err = channel.Publish(
		p.exchange,
		p.routingKey,
		false,
		false,
		amqp.Publishing{
			Body: []byte(body),
		},
	)
	if err != nil {
		return fmt.Errorf("cannot publish: %w", err)
	}

	select {
	case confirmation := <-channel.NotifyPublish(make(chan amqp.Confirmation, 1)):
		if !confirmation.Ack {
			return errors.New("failed to deliver message to exchange/queue")
		}
	case <-time.After(p.timeout):
		return errors.New("publishing timed out")
	}

	return nil
}
