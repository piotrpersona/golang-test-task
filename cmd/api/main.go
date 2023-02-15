package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"twitch_chat_analysis/internal/pubsub"

	"github.com/gin-gonic/gin"
)

func exit(err error) {
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
}

// POST body: { sender: String, receiver: String, message: String }
type messageRequest struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Message  string `json:"message"`
}

type messageResponse struct {
	Message string
}

func main() {
	r := gin.Default()

	publisher, err := pubsub.NewPublisher("")
	exit(err)

	r.POST("/message", func(c *gin.Context) {
		ctx := c.Request.Context()

		req := &messageRequest{}
		err := json.NewDecoder(c.Request.Body).Decode(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, messageResponse{Message: "cannot parse request body"})
			return
		}

		err = publisher.Publish(ctx, &pubsub.Message{
			Sender:   req.Sender,
			Receiver: req.Receiver,
			Message:  req.Message,
		})
		if err != nil {
			log.Printf("cannot publish message, err: %s\n", err)
			c.JSON(http.StatusInternalServerError, messageResponse{Message: "cannot publish message"})
			return
		}

		c.JSON(http.StatusOK, messageResponse{Message: "published message successfully"})
	})

	r.Run()
}
