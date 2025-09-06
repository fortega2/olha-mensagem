package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/fortega2/real-time-chat/internal/repository"
	"github.com/gorilla/websocket"
)

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	queries   *repository.Queries
	send      chan []byte
	user      *User
	ChannelID int
}

const (
	maxMessageSize = 512
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (writeWait * 9) / 10
)

func newClient(hub *Hub, conn *websocket.Conn, queries *repository.Queries, user *User, channelID int) *Client {
	return &Client{
		hub:       hub,
		conn:      conn,
		queries:   queries,
		send:      make(chan []byte, 256),
		user:      user,
		ChannelID: channelID,
	}
}

func (c *Client) processClientMessages() {
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

		message := NewChatMessage(c.user, chatType, content, c.ChannelID)
		jsonMsg, err := json.Marshal(message)
		if err != nil {
			c.hub.logger.Error("Failed to marshal message", "error", err, "user", c.user)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = c.queries.CreateMessage(ctx, repository.CreateMessageParams{
			UserID:    int64(c.user.ID),
			ChannelID: int64(c.ChannelID),
			Content:   content,
		})
		if err != nil {
			c.hub.logger.Error("Failed to persist message", "error", err, "userId", c.user.ID, "channelId", c.ChannelID)
			continue
		}
		c.hub.logger.Debug("Message create and broadcast", "user", c.user.Username, "channelId", c.ChannelID, "message", content)

		c.hub.broadcast <- jsonMsg
	}
}

func (c *Client) handleBroadcastMessages() {
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
