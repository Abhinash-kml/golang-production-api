package realtime

import (
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
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
		send: make(chan *ClientMessage),
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
	c.conn.SetReadDeadline(time.Now().Add(PongDuration))

	c.conn.SetPongHandler(func(appData string) error {
		c.conn.SetReadDeadline(time.Now().Add(PongDuration))
		return nil
	})

	for {
		message := new(ClientMessage)
		err := c.conn.ReadJSON(&message)
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

			// Get a writer for next message
			writer, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			// Batch remaining messages in channel to save bandwidth
			for value := range c.send {
				writer.Write([]byte{'\n'})
				writer.Write(value) // TODO: Fix this
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
