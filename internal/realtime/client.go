package realtime

import (
	"encoding/json"
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

type Client struct {
	uid  string
	conn *websocket.Conn
	send chan *ClientMessage
	hub  *Hub
}

func NewClient(uid string, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		uid:  uid,
		conn: conn,
		send: make(chan *ClientMessage, 100),
		hub:  hub,
	}
}

func (c *Client) ReadIncoming() {
	// Read message from Client
	// Forward to hub's send channel, hub will handle the rest
	// i.e. if receiver is local then send locally else publish to pub sub

	// Unregister from hub and close connection after read finishes
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(1064)
	c.conn.SetReadDeadline(time.Now().Add(PongWait))

	c.conn.SetPongHandler(func(appData string) error {
		c.conn.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})

	for {
		message := new(ClientMessage)
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

	for {
		select {
		case message, ok := <-c.send:
			if !ok { // Channel closed by Hub on unregister
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			zap.L().Info("Client message", zap.String("connection uid", c.uid), zap.String("payload", message.Payload))

			// Get a writer for next message
			writer, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			encoder := json.NewEncoder(writer)
			err = encoder.Encode(*message)
			if err != nil {
				zap.L().Warn("Json encoding failed", zap.Any("Message", message), zap.Error(err))
			}

			// Batch remaining messages in channel to save bandwidth
			n := len(c.send)
			for range n {
				value := <-c.send
				writer.Write([]byte{'\n'})
				err = encoder.Encode(*value)
				if err != nil {
					break
				}
			}

			if err := writer.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
