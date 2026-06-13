package kv

import (
	"fmt"

	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/redis/go-redis/v9"
)

func NewClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%s", config.LoadEnvVars().KV_PORT),
		Password: "",
		DB:       0,
	})
}
