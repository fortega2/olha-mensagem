package websocket

import (
	"encoding/json"
	"fmt"

	"github.com/fortega2/real-time-chat/internal/logger"
)

const notificationBuffer = 10

type Hub struct {
	logger       logger.Logger
	clients      map[*Client]struct{}
	broadcast    chan []byte
	notification chan Notification
	register     chan *Client
	unregister   chan *Client
	shutdown     chan struct{}
}

type Notification struct {
	message   string
	channelID int
}

func NewNotification(message string, channelID int) Notification {
	return Notification{
		message:   message,
		channelID: channelID,
	}
}

func NewHub(l logger.Logger) *Hub {
	return &Hub{
		logger:       l,
		clients:      make(map[*Client]struct{}),
		broadcast:    make(chan []byte),
		notification: make(chan Notification, notificationBuffer),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		shutdown:     make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = struct{}{}
			h.logger.Debug("Client registered", "user", client.user, "total_clients", len(h.clients))
			h.notification <- NewNotification(fmt.Sprintf("%s has joined the chat", client.user.Username), client.ChannelID)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.logger.Debug("Client unregistered", "user", client.user, "total_clients", len(h.clients))
				h.notification <- NewNotification(fmt.Sprintf("%s has left the chat", client.user.Username), client.ChannelID)
			}
		case message := <-h.broadcast:
			h.broadcastToChannel(message)
		case note := <-h.notification:
			go h.sendNotificationMessage(note)
		case <-h.shutdown:
			return
		}
	}
}

func (h *Hub) Shutdown() {
	close(h.shutdown)

	for client := range h.clients {
		close(client.send)
		client.conn.Close()
		delete(h.clients, client)
	}

	close(h.broadcast)
	close(h.register)
	close(h.unregister)
}

func (h *Hub) broadcastToChannel(message []byte) {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		h.logger.Error("Failed to unmarshal message for channel filtering", "error", err)
		return
	}

	for client := range h.clients {
		if client.ChannelID == msg.ChannelID {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}

func (h *Hub) sendNotificationMessage(note Notification) {
	notificationMsg := NewNotificationMessage(note.message, note.channelID)
	jsonMsg, err := json.Marshal(notificationMsg)
	if err != nil {
		h.logger.Error("Failed to marshal notification message", "error", err)
		return
	}

	h.logger.Debug("Sending notification message", "message", note.message, "channelID", note.channelID)
	h.broadcast <- jsonMsg
}
