package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
)

// NewClient cria uma nova conexão com o Redis
func NewClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

// Publish envia uma mensagem para um canal específico
func Publish(rdb *redis.Client, channel string, payload interface{}) error {
	return rdb.Publish(context.Background(), channel, payload).Err()
}