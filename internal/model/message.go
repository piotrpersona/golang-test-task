package model

import "time"

type Message struct {
	Sender, Receiver, Message string
	Created                   time.Time
}
