package message

import (
	"github.com/bendbennett/go-bits/messaging/config"
	"github.com/streadway/amqp"
)

type GetChannel func() (*amqp.Channel, error)

func ConfigureQueue(getChannel GetChannel, conf config.Messaging) error {
	channel, err := getChannel()
	if err != nil {
		return err
	}
	defer channel.Close()

	err = channel.ExchangeDeclare(
		conf.Exchange.Name,
		conf.Exchange.Kind,
		conf.Exchange.Durable,
		conf.Exchange.AutoDelete,
		conf.Exchange.Internal,
		conf.Exchange.NoWait,
		conf.Exchange.Args,
	)
	if err != nil {
		return err
	}

	_, err = channel.QueueDeclare(
		conf.Queue.Name,
		conf.Queue.Durable,
		conf.Queue.AutoDelete,
		conf.Queue.Exclusive,
		conf.Queue.NoWait,
		conf.Queue.Args,
	)
	if err != nil {
		return err
	}

	err = channel.QueueBind(
		conf.QueueBind.Name,
		conf.QueueBind.Key,
		conf.QueueBind.Exchange,
		conf.QueueBind.NoWait,
		conf.QueueBind.Args,
	)
	if err != nil {
		return err
	}

	return nil
}
