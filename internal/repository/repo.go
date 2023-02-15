package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"twitch_chat_analysis/internal/model"

	"github.com/redis/go-redis/v9"
)

type Writer interface {
	Save(ctx context.Context, m model.Message) (err error)
}

type Getter interface {
	Get(ctx context.Context, sender, receiver string) (msgs []model.Message, err error)
}

func NewWriter(rdb *redis.Client) (w Writer) {
	w = &repo{
		rdb: rdb,
	}
	return
}

func NewGetter(rdb *redis.Client) (w Getter) {
	w = &repo{
		rdb: rdb,
	}
	return
}

type repo struct {
	rdb *redis.Client
}

func (r repo) key(sender, receiver string) (key string) {
	return fmt.Sprintf("%s_%s", sender, receiver)
}

func (r repo) Save(ctx context.Context, m model.Message) (err error) {
	msgs, err := r.Get(ctx, m.Sender, m.Receiver)
	if err != nil {
		return
	}
	msgs = append(msgs, m)
	bytes, err := json.Marshal(msgs)
	if err != nil {
		err = fmt.Errorf("failed to marshal message: %v", err)
		return
	}
	err = r.rdb.Set(ctx, r.key(m.Sender, m.Receiver), string(bytes), 0).Err()
	if err != nil {
		err = fmt.Errorf("cannot set, err: %w", err)
		return
	}
	return
}

func (r repo) Get(ctx context.Context, sender, receiver string) (msgs []model.Message, err error) {
	bytes, getErr := r.rdb.Get(ctx, r.key(sender, receiver)).Bytes()
	if getErr != nil {
		if errors.Is(getErr, redis.Nil) {
			return
		}
		err = fmt.Errorf("cannot get messages, err: %w", getErr)
		return
	}
	err = json.Unmarshal(bytes, &msgs)
	if err != nil {
		err = fmt.Errorf("cannot unmarshal messages, err: %w", err)
	}
	return
}
