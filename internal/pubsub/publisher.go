package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"twitch_chat_analysis/internal/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

const messageQ = "messages"

type Publisher interface {
	Publish(ctx context.Context, m *model.Message) (err error)
}

type publisher struct {
	channel *amqp.Channel
}

func connect(uri string) (conn *amqp.Connection, channel *amqp.Channel, err error) {
	conn, err = amqp.Dial(uri)
	if err != nil {
		err = fmt.Errorf("cannot connect to amqp, err: %w", err)
		return
	}

	channel, err = conn.Channel()
	if err != nil {
		err = fmt.Errorf("cannot get channel, err: %w", err)
		return
	}
	return
}

func NewPublisher(uri string) (p Publisher, shutdown func(), err error) {
	conn, channel, err := connect(uri)
	if err != nil {
		return
	}

	_, err = channel.QueueDeclare(
		messageQ, // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		err = fmt.Errorf("cannot decalre Q, err: %w", err)
		return
	}

	p = &publisher{
		channel: channel,
	}
	shutdown = func() {
		conn.Close()
	}
	return
}

func (p publisher) Publish(ctx context.Context, m *model.Message) (err error) {
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
