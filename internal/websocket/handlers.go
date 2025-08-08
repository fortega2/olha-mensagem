package websocket

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

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

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user := GetUserByID(userId)
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not upgrade connection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := newClient(hub, conn, user)
	client.hub.register <- client

	go client.readHubMessages()
	go client.processClientMessages()
}

func Shutdown() {
	hub.Shutdown()
}
