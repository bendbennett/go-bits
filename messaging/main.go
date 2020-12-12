package main

import (
	"context"
	"github.com/bendbennett/go-bits/concurrency/rabbit/config"
	"github.com/bendbennett/go-bits/concurrency/rabbit/message"
	"log"
	"strconv"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf := config.New()

	connManager := message.NewConnManager(conf.Messaging.Connection)
	err := connManager.Connect(ctx)
	if err != nil {
		log.Panicf("cannot connect at boot: %v", err)
	}
	defer connManager.Close()

	err = message.ConfigureQueue(
		connManager.Channel,
		conf.Messaging,
	)
	if err != nil {
		log.Panicf("cannot configure queue: %v", err)
	}

	publisher := message.NewPublisher(
		connManager.Channel,
		conf.Messaging.Exchange.Name,
		conf.Messaging.QueueBind.Key,
		conf.Messaging.Publish.Timeout,
	)

	for i := 0; i < 100; i++ {
		body := "Hello World! - " + strconv.Itoa(i)
		err = publisher.Publish(body)
		if err != nil {
			log.Println(err)
		}
		log.Printf("Sent %s", body)
		time.Sleep(1 * time.Second)
	}
}
