package websocket

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
	user *User
}

const (
	maxMessageSize = 512
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (writeWait * 9) / 10
)

func newClient(hub *Hub, conn *websocket.Conn, user *User) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
		user: user,
	}
}

func (c *Client) readClientMessages() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msgBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
				c.hub.logger.Debug("Client disconnected", "reason", err.Error(), "user", c.user)
			} else {
				c.hub.logger.Error("Error reading message", "error", err, "user", c.user)
			}
			break
		}

		content := string(bytes.TrimSpace(bytes.ReplaceAll(msgBytes, []byte{'\n'}, []byte{' '})))
		if content == "" {
			continue
		}

		c.hub.logger.Debug("Message received", "user", c.user, "content", content)

		message := newChatMessage(c.user, chatType, content)
		jsonMsg, err := json.Marshal(message)
		if err != nil {
			c.hub.logger.Error("Failed to marshal message", "error", err, "user", c.user)
			continue
		}

		c.hub.broadcast <- jsonMsg
	}
}

func (c *Client) writeClientMessages() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				c.hub.logger.Debug("Send channel closed, closing WS", "user", c.user.Username)
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.hub.logger.Debug("Delivering message to client", "user", c.user.Username, "message", string(message))

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.hub.logger.Error("Failed to write message", "user", c.user.Username, "error", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.hub.logger.Error("Ping failed", "user", c.user.Username, "error", err)
				return
			}
		}
	}
}
