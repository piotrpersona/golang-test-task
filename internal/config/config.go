package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Specification struct {
	RedisAddr  string `default:"localhost:6379"`
	RabbitConn string `default:"amqp://user:password@localhost:7001/"`
}

func Load() (s Specification) {
	err := envconfig.Process("app", &s)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}
