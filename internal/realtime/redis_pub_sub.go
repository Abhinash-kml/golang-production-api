package realtime

import (
	"context"

	"github.com/abhinash-kml/go-api-server/internal/connections"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisPubSub struct {
	rdb          *redis.Client
	pubsub       *redis.PubSub
	incomingChan <-chan *redis.Message
}

func NewRedisPubSub(conn *connections.RedisConnection) RedisPubSub {
	return RedisPubSub{rdb: conn.Client}
}

func (r *RedisPubSub) Initialize() error {
	ctx := context.Background()
	r.pubsub = r.rdb.Subscribe(ctx)

	return nil
}

// TODO: Check for failed publish
func (r *RedisPubSub) Publish(channel string, messsage *Envelope) error {
	ctx := context.Background()
	r.rdb.Publish(ctx, channel, messsage)
	return nil
}

func (r *RedisPubSub) Subscribe(channel string) {
	if r.rdb == nil {
		zap.L().Fatal("Redis pub-sub subcribe failed due to nil client pointer")
	}

	if r.pubsub == nil {
		// zap.L().Fatal("Redis pub-sub subscribe failed due to nil pubsub pointer")
		ctx := context.Background()
		r.pubsub = r.rdb.Subscribe(ctx, channel)
		r.incomingChan = r.pubsub.Channel()
	} else {
		ctx := context.Background()
		err := r.pubsub.Subscribe(ctx, channel)
		if err != nil {
			zap.L().Warn("Subcribe to channel failed", zap.String("channel", channel), zap.Error(err))
			return
		}
	}
}

// INFO: Subjected to improvement
func (r *RedisPubSub) ListenToSubscriptions() <-chan any {
	out := make(chan any, 100)
	go func() {
		for message := range r.incomingChan {
			out <- message
		}
	}()

	return out
}
