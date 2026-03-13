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

const broadcastChannelString = "@"

type Hub struct {
	register   chan *Client
	unregister chan *Client
	send       chan *Envelope
	broadcast  chan *Envelope
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
		send:       make(chan *Envelope, 100),
		broadcast:  make(chan *Envelope, 100),
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

func (h *Hub) Subscribe(uid string) {
	h.subscribe <- uid
}

func (h *Hub) Broadcast(message *Envelope) {
	h.broadcast <- message
}

func (h *Hub) Stop() {
	h.cancel()
}

func (h *Hub) Run() {
	// Subscribe to special uid - @, for internode broadcast message
	h.Subscribe(broadcastChannelString)

	for {
		select {
		case client := <-h.register:
			h.HandleRegistration(client)
		case client := <-h.unregister:
			h.HandleUnregistration(client)
		case broadcastMessage := <-h.broadcast:
			h.HandleBroadcast(broadcastMessage)
		case message := <-h.send: // TODO: Use worker pool here
			h.HandleClientMessages(message)
		case subscription := <-h.subscribe:
			h.HandleSubscribtionRequests(subscription)
		}
	}
}

func (h *Hub) HandleRegistration(c *Client) {
	h.store.Add(c.uid, c)
	zap.L().Debug("Websocket client connected", zap.String("uid", c.uid))
}

func (h *Hub) HandleUnregistration(c *Client) {
	close(c.send)
	h.store.Remove(c.uid)
	zap.L().Debug("Websocket client disconnected", zap.String("uid", c.uid))
}

func (h *Hub) HandleClientMessages(message *Envelope) {
	zap.L().Debug("Websocket Message", zap.String("from", message.Header.SenderID), zap.String("to", message.Header.RecieverID), zap.String("payload", string(message.Data)))

	// If the message is broadcast then send it to broadcast channel
	if message.Header.RecieverID == broadcastChannelString {
		h.broadcast <- message
		return
	}

	// If reciever is present locally on this node send directly
	if user := h.store.Get(message.Header.RecieverID); user != nil {
		user.send <- message
		return
	}

	// Else
	// Publish to Pub Sub so other nodes can handle from there
	h.pubsub.Publish(message.Header.RecieverID, message)
}

func (h *Hub) HandleBroadcast(message *Envelope) {
	// Send to local clients
	h.store.ForEach(func(c *Client) {
		c.send <- message
	})

	// Send to pub-sub broadcast channel
	err := h.pubsub.Publish(broadcastChannelString, message)
	if err != nil {
		zap.L().Info("PubSub broadcast failed", zap.Error(err))
	}
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
				internal := new(Envelope)
				if err := json.Unmarshal([]byte(redisMsg.Payload), internal); err != nil {
					zap.L().Error("Unmarshal failed", zap.Error(err))
					continue
				}

				zap.L().Debug("Redis channel message", zap.String("from", internal.Header.SenderID), zap.String("to", internal.Header.RecieverID), zap.String("payload", string(internal.Data)))

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
