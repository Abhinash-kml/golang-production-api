package realtime

import (
	"context"
)

type Hub struct {
	register   chan *Client
	unregister chan *Client
	send       chan *ClientMessage
	broadcast  chan *ClientMessage

	store  ISessionStore
	pubsub IPubSub

	// Mutex only needed if store doesn't provide internal concurrency
	// mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewHub(store *ISessionStore, pubsub *IPubSub) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	hub := &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		send:       make(chan *ClientMessage),
		broadcast:  make(chan *ClientMessage),
		store:      *store,
		pubsub:     *pubsub,
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

func (h *Hub) Stop() {
	h.cancel()
}
