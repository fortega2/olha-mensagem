package websocket

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/fortega2/real-time-chat/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

type WebsocketHandler struct {
	logger  logger.Logger
	queries *repository.Queries
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	hub  = NewHub()
	once sync.Once
)

func init() {
	once.Do(func() {
		go hub.Run()
	})
}

func NewWebsocketHandler(l logger.Logger, q *repository.Queries) *WebsocketHandler {
	return &WebsocketHandler{
		logger:  l,
		queries: q,
	}
}

func (wh *WebsocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		wh.logger.Error("Invalid user ID", "error", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	wh.logger.Info("WebSocket connection attempt", "userID", userId)

	dbUser, err := wh.queries.GetUserByID(r.Context(), int64(userId))
	if err != nil {
		wh.logger.Error("Failed to get user by ID", "error", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	wh.logger.Debug("User found. Attempting to upgrade to WebSocket", "userID", userId, "username", dbUser.Username)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		wh.logger.Error("Failed to upgrade connection", "error", err)
		http.Error(w, "Could not upgrade connection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	wh.logger.Info("WebSocket connection established", "userID", userId, "username", dbUser.Username)

	user := NewUser(int(dbUser.ID), dbUser.Username)
	client := newClient(hub, conn, user)

	wh.logger.Debug("Registering new client to hub", "user", user)

	client.hub.register <- client

	go client.readHubMessages()
	go client.processClientMessages()
}

func Shutdown() {
	hub.Shutdown()
}
