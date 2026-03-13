package realtime

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/google/uuid"
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
	register          chan *Client
	unregister        chan *Client
	send              chan *Envelope
	broadcast         chan *Envelope
	incomingBroadcast chan *Envelope
	subscribe         chan string

	store      ISessionStore
	pubsub     IPubSub
	pubsubtype int

	// Mutex only needed if store doesn't provide internal concurrency
	// mu     sync.RWMutex
	once   sync.Once
	ctx    context.Context
	cancel context.CancelFunc

	nodeID uuid.UUID
}

func NewHub(store ISessionStore, pubsub IPubSub, pubsubtype int) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	uuid, _ := uuid.NewV7()
	hub := &Hub{
		register:          make(chan *Client, 100),
		unregister:        make(chan *Client, 100),
		send:              make(chan *Envelope, 100),
		broadcast:         make(chan *Envelope, 100),
		incomingBroadcast: make(chan *Envelope, 100),
		subscribe:         make(chan string, 100),
		store:             store,
		pubsub:            pubsub,
		pubsubtype:        pubsubtype,
		ctx:               ctx,
		cancel:            cancel,
		nodeID:            uuid,
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
		case incomingBroadcastMessage := <-h.incomingBroadcast:
			h.HandleIncomingBroadcast(incomingBroadcastMessage)
		case message := <-h.send: // TODO: Use worker pool here
			h.HandleClientMessages(message)
		case subscription := <-h.subscribe:
			h.HandleSubscribtionRequests(subscription)
		}
	}
}

func (h *Hub) SetMessageMetadata(message *Envelope) {
	message.Header.SenderID = h.nodeID.String()
	message.Header.Hops++
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

	h.SetMessageMetadata(message)

	// TODO: Improve this
	// If the message is broadcast then send it to broadcast channel
	if message.Header.Category == CategoryBroadcast {
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

func (h *Hub) HandleIncomingBroadcast(message *Envelope) {
	// Broadcast to local connections
	h.store.ForEach(func(c *Client) {
		c.send <- message
	})
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

			switch message := msg.(type) {
			case *redis.Message:
				h.HandleRedisPubSubMessage(message)
			case nil:
				break
			}
		}
	}
}

func (h *Hub) ProcessIncomingBroadcast(message *Envelope) {
	// Prevent message loop
	// Check for broadcast message initiated from this node and discard them as they are already processed by local broadcast
	if message.Header.SenderID == h.nodeID.String() {
		zap.L().Debug("Dropped broadcast echo", zap.String("messsageID", message.Header.CorrelationID))
		return
	}

	// Drop messages which have hopped between nodes N times
	if message.Header.Hops >= 2 {
		zap.L().Debug("Dropped broadcast echo due to hops", zap.String("messsageID", message.Header.CorrelationID), zap.Int("hops", message.Header.Hops))
		return
	}

	// TODO: Check seen messages from small cache to prevent duplication

	// Send incoming broadcast to incoming broadcast channel and skin the current iteration of loop
	h.incomingBroadcast <- message
}

func (h *Hub) HandleRedisPubSubMessage(redisMessage *redis.Message) {
	message := new(Envelope)
	if err := json.Unmarshal([]byte(redisMessage.Payload), message); err != nil {
		zap.L().Error("Unmarshal failed", zap.Error(err))
		return
	}

	zap.L().Debug("Redis channel message", zap.String("from", message.Header.SenderID), zap.String("to", message.Header.RecieverID), zap.String("payload", string(message.Data)))

	// If incoming message is a broadcast then process it first, if success then it will be forwarded to incoming broadcast handler
	if message.Header.Category == CategoryBroadcast {
		h.ProcessIncomingBroadcast(message)
		return
	}

	select {
	case h.send <- message:
		// Success: Message sent to Hub
	default:
		// Failure: Hub's buffer is full.
		// We drop the message to keep the Redis consumer alive.
		// Or we can retry later
		zap.L().Warn("Hub busy: dropping Redis message", zap.Any("message", message))
	}
}

func (h *Hub) HandleNatsPubSubMessage(message *Envelope) {

}
