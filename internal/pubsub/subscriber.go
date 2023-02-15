package pubsub

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SubCallback func(ctx context.Context, m amqp.Delivery) error

type Subscriber interface {
	Subscribe(ctx context.Context, handle SubCallback) (err error)
}

type subscriber struct {
	channel *amqp.Channel
}

func NewSubscriber(uri string) (s Subscriber, shutdown func(), err error) {
	uri = "amqp://user:password@localhost:7001/"
	conn, channel, err := connect(uri)
	if err != nil {
		return
	}
	s = &subscriber{
		channel: channel,
	}
	shutdown = func() {
		conn.Close()
	}
	return
}

func (s *subscriber) Subscribe(ctx context.Context, handle SubCallback) (err error) {
	msgs, err := s.channel.Consume(
		messageQ, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		err = fmt.Errorf("cannot consume, err: %w", err)
		return
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, more := <-msgs:
				if !more {
					return
				}
				err = handle(ctx, msg)
				if err != nil {
					log.Printf("error while processing message: %s", err)
				}
			}
		}
	}()

	return
}
