package realtime

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisPubSub struct {
	rdb          *redis.Client
	pubsub       *redis.PubSub
	incomingChan <-chan *redis.Message
}

func NewRedisPubSub(client *redis.Client) RedisPubSub {
	return RedisPubSub{rdb: client}
}

func (r *RedisPubSub) Initialize() error {
	ctx := context.Background()
	r.pubsub = r.rdb.Subscribe(ctx)
	iface, err := r.pubsub.Receive(ctx)
	if err != nil {
		fmt.Println("Failed to recieve message from redis pubsub handle")
		return err
	}

	switch iface.(type) {
	case *redis.Subscription:
		fmt.Println("Subscribtion message from pubsub")
	case *redis.Message:
		fmt.Println("Message from pubsub")
	case *redis.Pong:
		fmt.Println("Pong from redis pubsub")
	}

	r.incomingChan = r.pubsub.Channel()
	return nil
}

// TODO: Check for failed publish
func (r *RedisPubSub) Publish(channel string, messsage *ClientMessage) error {
	ctx := context.Background()
	r.rdb.Publish(ctx, channel, messsage)
	return nil
}

func (r *RedisPubSub) Subscribe(channel string) {
	ctx := context.Background()
	r.pubsub.Subscribe(ctx, channel)
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
