package connections

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func (c *Redis) Connect(option *redis.Options) error {
	rdb := redis.NewClient(option)
	result, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	if result != "Pong" {
		return errors.New("Redis Ping failed")
	}

	c.Client = rdb

	return nil
}

func (c *Redis) HealthCheck() bool {
	if err := c.Client.Ping(context.Background()).Err(); err != nil {
		return false
	}

	return true
}
