package cache

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

var Ctx = context.Background()

func ConnectRedis() error {
	Client = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	return Client.Ping(Ctx).Err()
}
