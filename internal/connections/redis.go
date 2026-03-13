package connections

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisConnection struct {
	option *redis.Options
	Client *redis.Client
}

func NewRedisConnection(options *redis.Options) *RedisConnection {
	connection := &RedisConnection{option: options}
	err := connection.Connect()
	if err != nil {
		zap.L().Fatal("Redis connection failed", zap.Error(err))
		return nil
	}

	return connection
}

func (c *RedisConnection) Connect() error {
	rdb := redis.NewClient(c.option)
	result, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	if result != "PONG" {
		return errors.New("Redis Ping failed")
	}

	c.Client = rdb

	c.onConnnect()

	return nil
}

func (c *RedisConnection) onConnnect() {
	zap.L().Info("Connected to redis database", zap.String("address", c.option.Addr), zap.Int("db", c.option.DB))
}

func (c *RedisConnection) HealthCheck() bool {
	if err := c.Client.Ping(context.Background()).Err(); err != nil {
		return false
	}

	return true
}
