package realtime

import (
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	// Maximum time to wait for a write operation
	WriteWait = time.Second * 10
	// Maxium time to wait for a pong response
	PongWait = time.Second * 60
	// Interval for sending ping message (must be less than pong wait)
	PingInterval = (PongWait * 9) / 10
	// Maximum message size
	MaxMessageSize = 512 * 1024
)

type ConnectionStats struct {
	ConnectedAt      time.Time
	LastPingAt       time.Time
	LastPongAt       time.Time
	MessagesSend     int64
	MessagesReceived int64
	PingsSent        int64
	PongsReceived    int64
}

type Client struct {
	uid   string
	conn  *websocket.Conn
	send  chan *Envelope
	hub   *Hub
	stats ConnectionStats
}

func NewClient(uid string, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		uid:  uid,
		conn: conn,
		send: make(chan *Envelope, 100),
		hub:  hub,
		stats: ConnectionStats{
			ConnectedAt: time.Now(),
		},
	}
}

func (c *Client) ReadIncoming() {
	// Read message from Client
	// Forward to hub's send channel, hub will handle the rest
	// i.e. if receiver is local then send locally else publish to pub sub

	// Unregister from hub and close connection after read finishes
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(1064)
	c.conn.SetReadDeadline(time.Now().Add(PongWait))

	c.conn.SetPongHandler(func(appData string) error {
		c.conn.SetReadDeadline(time.Now().Add(PongWait))
		c.RecordLastPong()
		return nil
	})

	for {
		message := new(Envelope)
		message.Header.SourceID = c.uid

		err := c.conn.ReadJSON(message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
				websocket.CloseNormalClosure) {
				zap.L().Info("Websocket Read Fail", zap.Error(err))
			}
			break
		}

		c.hub.send <- message
	}
}

func (c *Client) WriteOutgoing() {
	// Ticker for periodic ping message
	ticker := time.NewTicker(PingInterval)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.send:
			if !ok { // Channel closed by Hub on unregister
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.SetWriteDeadline(time.Now().Add(WriteWait)) // Failing to set this will make the connection currupt
			zap.L().Debug("Client message", zap.String("connection uid", c.uid), zap.String("payload", string(message.Data)))

			// Get a writer for next message
			writer, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			encoder := json.NewEncoder(writer)
			err = encoder.Encode(message)
			if err != nil {
				zap.L().Warn("Json encoding failed", zap.Any("Message", message), zap.Error(err))
			}

			// Batch remaining messages in channel to save bandwidth
			n := len(c.send)
			for range n {
				value := <-c.send
				writer.Write([]byte{'\n'})
				err = encoder.Encode(value)
				if err != nil {
					break
				}
			}

			if err := writer.Close(); err != nil {
				zap.L().Info("Websocket write failed", zap.Error(err))
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				zap.L().Info("Websocket write failed in ticker", zap.Error(err))
				return
			}
			c.RecordLastPing()
		}
	}
}

func (c *Client) RecordMessageSent() {
	atomic.AddInt64(&c.stats.MessagesSend, 1)
}

func (c *Client) RecordMessageReceived() {
	atomic.AddInt64(&c.stats.MessagesReceived, 1)
}

func (c *Client) RecordLastPong() {
	atomic.AddInt64(&c.stats.PongsReceived, 1)
	c.stats.LastPongAt = time.Now()
}

func (c *Client) RecordLastPing() {
	atomic.AddInt64(&c.stats.PingsSent, 1)
	c.stats.LastPingAt = time.Now()
}

func (c *Client) GetStats() ConnectionStats {
	return ConnectionStats{
		ConnectedAt:      c.stats.ConnectedAt,
		LastPingAt:       c.stats.LastPingAt,
		LastPongAt:       c.stats.LastPongAt,
		MessagesSend:     atomic.LoadInt64(&c.stats.MessagesSend),
		MessagesReceived: atomic.LoadInt64(&c.stats.MessagesReceived),
		PingsSent:        atomic.LoadInt64(&c.stats.PingsSent),
		PongsReceived:    atomic.LoadInt64(&c.stats.PongsReceived),
	}
}

func (c *Client) Latency() time.Duration {
	if c.stats.LastPongAt.IsZero() || c.stats.LastPingAt.IsZero() {
		return 0
	}

	return c.stats.LastPongAt.Sub(c.stats.LastPingAt)
}
