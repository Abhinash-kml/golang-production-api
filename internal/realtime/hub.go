package realtime

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
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
	ctx    context.Context
	cancel context.CancelFunc
}

func NewHub(store *ISessionStore, pubsub *IPubSub, pubsubtype int) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	hub := &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		send:       make(chan *ClientMessage),
		broadcast:  make(chan *ClientMessage),
		subscribe:  make(chan string),
		store:      *store,
		pubsub:     *pubsub,
		pubsubtype: pubsubtype,
		ctx:        ctx,
		cancel:     cancel,
	}

	return hub
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Run() {
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
	// If reciever is present locally on this node send directly
	if user := h.store.Get(message.ReceiverID); user != nil {
		user.send <- message
		return
	}

	// Else
	// Publish to Pub Sub so other nodes can handle from there
	h.pubsub.Publish(message.ReceiverID, message)
}

func (h *Hub) HandleSubscribtionRequests(subscription string) {
	h.pubsub.Subscribe(subscription)
}

// TODO: Improve this
func (h *Hub) ListenToSubscriptions() {
	incomingMessage := h.pubsub.ListenToSubscriptions()
	for message := range incomingMessage {
		switch value := message.(type) {
		case *redis.Message:
			internal := new(ClientMessage)
			json.Unmarshal([]byte(value.Payload), internal)
			h.send <- internal
		}
	}
}

func (h *Hub) Stop() {
	h.cancel()
}
