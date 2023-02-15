package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const messageQ = "messages"

type Message struct {
	Sender, Receiver, Message string
}

type Publisher interface {
	Publish(ctx context.Context, m *Message) (err error)
}

type publisher struct {
	channel *amqp.Channel
}

func NewPublisher(uri string) (p Publisher, err error) {
	uri = "amqp://user:password@localhost:7001/"
	conn, err := amqp.Dial(uri)
	if err != nil {
		err = fmt.Errorf("cannot connect to amqp, err: %w", err)
		return
	}
	// defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		err = fmt.Errorf("cannot get channel, err: %w", err)
		return
	}

	p = &publisher{
		channel: channel,
	}
	return
}

func (p publisher) Publish(ctx context.Context, m *Message) (err error) {
	body, err := json.Marshal(m)
	if err != nil {
		err = fmt.Errorf("cannot marshal message, err: %w", err)
		return
	}
	err = p.channel.PublishWithContext(ctx, "", messageQ, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         body,
	})
	if err != nil {
		err = fmt.Errorf("cannot publish message, err: %w", err)
	}
	return
}
