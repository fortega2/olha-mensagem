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
	hub  *Hub
	once sync.Once
)

func NewWebsocketHandler(l logger.Logger, q *repository.Queries) *WebsocketHandler {
	once.Do(func() {
		hub = NewHub(l)
		go hub.Run()
	})

	return &WebsocketHandler{
		logger:  l,
		queries: q,
	}
}

func (wh *WebsocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userId, ok := wh.getUserIDFromRequest(r, w)
	if !ok {
		return
	}

	channelId, ok := wh.getChannelIDFromRequest(r, w)
	if !ok {
		return
	}

	wh.logger.Debug("WebSocket connection attempt", "channelID", channelId, "userID", userId)

	dbUser, err := wh.queries.GetUserByID(r.Context(), int64(userId))
	if err != nil {
		wh.logger.Error("Failed to get user by ID", "error", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		wh.logger.Error("Failed to upgrade connection", "error", err)
		http.Error(w, "Could not upgrade connection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	wh.logger.Info("WebSocket connection established", "userID", userId, "username", dbUser.Username)

	user := NewUser(int(dbUser.ID), dbUser.Username)
	client := newClient(hub, conn, wh.queries, user, channelId)

	client.hub.register <- client

	go client.handleBroadcastMessages()
	go client.processClientMessages()
}

func (wh *WebsocketHandler) getUserIDFromRequest(r *http.Request, w http.ResponseWriter) (int, bool) {
	userIdStr := chi.URLParam(r, "userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		wh.logger.Error("Invalid user ID", "error", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return 0, false
	}
	return userId, true
}

func (wh *WebsocketHandler) getChannelIDFromRequest(r *http.Request, w http.ResponseWriter) (int, bool) {
	channelIdStr := chi.URLParam(r, "channelId")
	channelId, err := strconv.Atoi(channelIdStr)
	if err != nil {
		wh.logger.Error("Invalid channel ID", "error", err)
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return 0, false
	}
	return channelId, true
}

func Shutdown() {
	hub.Shutdown()
}
