package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func NewClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

func Publish(rdb *redis.Client, channel string, payload interface{}) error {
	return rdb.Publish(ctx, channel, payload).Err()
}