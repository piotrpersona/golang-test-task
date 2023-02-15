package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"twitch_chat_analysis/internal/config"
	"twitch_chat_analysis/internal/model"
	"twitch_chat_analysis/internal/pubsub"
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

// TODO: Read DB & MQ connection from ENV
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Load()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: "",
		DB:       0,
	})
	exit(rdb.Ping(ctx).Err())

	handle := messageHandler(rdb)

	s, shutdown, err := pubsub.NewSubscriber(cfg.RabbitConn)
	exit(err)
	defer shutdown()

	err = s.Subscribe(ctx, handle)
	exit(err)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	log.Printf("waiting for messages. To exit press CTRL+C")
	sig := <-sigc
	cancel()
	log.Printf("received signal: %s, terminating application", sig.String())
}
