package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"sort"

	"twitch_chat_analysis/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func exit(err error) {
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
}

type messageResponse struct {
	Message string
}

func main() {
	mainCtx := context.Background()

	r := gin.Default()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	exit(rdb.Ping(mainCtx).Err())

	getter := repository.NewGetter(rdb)

	r.GET("/message/list", func(c *gin.Context) {
		ctx := c.Request.Context()

		sender := c.Query("sender")
		receiver := c.Query("receiver")

		if sender == "" {
			c.JSON(http.StatusBadRequest, messageResponse{Message: "sender empty"})
		}
		if receiver == "" {
			c.JSON(http.StatusBadRequest, messageResponse{Message: "receiver empty"})
		}

		msgs, err := getter.Get(ctx, sender, receiver)
		if err != nil {
			log.Printf("cannot get messages, err: %s\n", err)
			c.JSON(http.StatusInternalServerError, messageResponse{Message: "cannot get messages"})
			return
		}
		sort.Slice(msgs[:], func(i, j int) bool {
			return msgs[i].Created.After(msgs[j].Created)
		})

		c.JSON(http.StatusOK, msgs)
	})

	r.Run(":8001")
}
