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
	notification chan string
	register     chan *Client
	unregister   chan *Client
	shutdown     chan struct{}
}

func NewHub(l logger.Logger) *Hub {
	return &Hub{
		logger:       l,
		clients:      make(map[*Client]struct{}),
		broadcast:    make(chan []byte),
		notification: make(chan string, notificationBuffer),
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
			h.notification <- fmt.Sprintf("%s has joined the chat", client.user.Username)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.logger.Debug("Client unregistered", "user", client.user, "total_clients", len(h.clients))
				h.notification <- fmt.Sprintf("%s has left the chat", client.user.Username)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
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

func (h *Hub) sendNotificationMessage(message string) {
	notificationMsg := newNotificationMessage(message)
	jsonMsg, err := json.Marshal(notificationMsg)
	if err != nil {
		h.logger.Error("Failed to marshal notification message", "error", err)
		return
	}

	h.logger.Debug("Sending notification message", "message", message)
	h.broadcast <- jsonMsg
}
