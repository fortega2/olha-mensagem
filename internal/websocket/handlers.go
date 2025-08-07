package websocket

import (
	"net/http"
	"sync"

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
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not upgrade connection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := NewClient(hub, conn)
	client.hub.register <- client

	go client.readHubMessages()
	go client.processClientMessages()
}

func Shutdown() {
	hub.Shutdown()
}
