package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"twitch_chat_analysis/internal/model"
	"twitch_chat_analysis/internal/repository"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

func exit(err error) {
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
}

func messageHandler(rdb *redis.Client) func(ctx context.Context, m amqp.Delivery) error {
	writer := repository.NewWriter(rdb)
	return func(ctx context.Context, m amqp.Delivery) (err error) {
		var msg model.Message
		err = json.Unmarshal(m.Body, &msg)
		if err != nil {
			err = fmt.Errorf("cannot unmarshal message, err: %w", err)
			return
		}
		err = writer.Save(ctx, msg)
		if err != nil {
			err = fmt.Errorf("cannot save message, err: %w", err)
			return
		}
		err = m.Ack(false)
		if err != nil {
			err = fmt.Errorf("cannot ack")
			return
		}
		return
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	exit(rdb.Ping(ctx).Err())

	handle := messageHandler(rdb)

	uri := "amqp://user:password@localhost:7001/"
	conn, err := amqp.Dial(uri)
	exit(err)

	channel, err := conn.Channel()
	exit(err)

	msgs, err := channel.Consume(
		"messages", // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	exit(err)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

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

	log.Printf("waiting for messages. To exit press CTRL+C")
	sig := <-sigc
	cancel()
	log.Printf("received signal: %s, terminating application", sig.String())
}
