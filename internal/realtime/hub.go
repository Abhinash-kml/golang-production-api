package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	PubSubTypeNone = iota
	PubSubTypeMemory
	PubSubTypeRedis
	PubSubTypeNats
	PubSubTypeRabbitMQ
)

type Hub struct {
	register   chan *Client
	unregister chan *Client
	send       chan *ClientMessage
	broadcast  chan *ClientMessage
	subscribe  chan string

	store      ISessionStore
	pubsub     IPubSub
	pubsubtype int

	// Mutex only needed if store doesn't provide internal concurrency
	// mu     sync.RWMutex
	once   sync.Once
	ctx    context.Context
	cancel context.CancelFunc
}

func NewHub(store ISessionStore, pubsub IPubSub, pubsubtype int) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	hub := &Hub{
		register:   make(chan *Client, 100),
		unregister: make(chan *Client, 100),
		send:       make(chan *ClientMessage, 100),
		broadcast:  make(chan *ClientMessage, 100),
		subscribe:  make(chan string, 100),
		store:      store,
		pubsub:     pubsub,
		pubsubtype: pubsubtype,
		ctx:        ctx,
		cancel:     cancel,
	}

	return hub
}

func (h *Hub) Initialize() {
	err := h.pubsub.Initialize()
	if err != nil {
		zap.L().Fatal("Hub initialize failed", zap.Error(err))
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Run() {
	// go h.ListenToSubscriptions()

	for {
		select {
		case client := <-h.register:
			h.store.Add(client.uid, client)
		case client := <-h.unregister:
			close(client.send)
			h.store.Remove(client.uid)
		case broadcastMessage := <-h.broadcast:
			h.store.ForEach(func(client *Client) {
				client.send <- broadcastMessage
			})
		case message := <-h.send: // TODO: Use worker pool here
			h.HandleClientMessages(message)
		case subscription := <-h.subscribe:
			h.HandleSubscribtionRequests(subscription)
		}
	}
}

func (h *Hub) HandleClientMessages(message *ClientMessage) {
	zap.L().Info("Websocket Message", zap.String("from", message.SenderID), zap.String("to", message.ReceiverID), zap.String("payload", message.Payload))

	// If reciever is present locally on this node send directly
	if user := h.store.Get(message.ReceiverID); user != nil {
		user.send <- message
		return
	}

	// Else
	// Publish to Pub Sub so other nodes can handle from there
	h.pubsub.Publish(message.ReceiverID, message)
}

func (h *Hub) Subscribe(uid string) {
	h.subscribe <- uid
}

func (h *Hub) HandleSubscribtionRequests(subscription string) {
	h.pubsub.Subscribe(subscription)
	h.once.Do(func() {
		go h.ListenToSubscriptions()
	})
}

// TODO: Improve this
func (h *Hub) ListenToSubscriptions() {
	incomingMessage := h.pubsub.ListenToSubscriptions()
	for {
		select {
		case <-h.ctx.Done():
			return // Graceful shutdown
		case msg, ok := <-incomingMessage:
			if !ok {
				return // Channel closed
			}

			fmt.Println(msg)

			if redisMsg, ok := msg.(*redis.Message); ok {
				internal := new(ClientMessage)
				if err := json.Unmarshal([]byte(redisMsg.Payload), internal); err != nil {
					zap.L().Error("Unmarshal failed", zap.Error(err))
					continue
				}

				zap.L().Debug("Redis channel message", zap.String("from", internal.SenderID), zap.String("to", internal.ReceiverID), zap.String("payload", internal.Payload))

				// --- NON-BLOCKING SEND START ---
				select {
				case h.send <- internal:
					// Success: Message sent to Hub
				default:
					// Failure: Hub's buffer is full.
					// We drop the message to keep the Redis consumer alive.
					zap.L().Warn("Hub busy: dropping Redis message",
						zap.String("payload", redisMsg.Payload))
				}
				// --- NON-BLOCKING SEND END ---
			}
		}
	}
}

func (h *Hub) Broadcast(message *ClientMessage) {
	h.broadcast <- message
}

func (h *Hub) Stop() {
	h.cancel()
}
