package server

import (
	"net/http"

	"github.com/fortega2/real-time-chat/internal/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Start() error {
	r := chi.NewRouter()

	configMiddlewares(r)
	setRoutes(r)

	return http.ListenAndServe(":8080", r)
}

func configMiddlewares(r *chi.Mux) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
}

func setRoutes(r *chi.Mux) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "internal/templates/index.html")
	})
	r.Get("/ws", websocket.HandleWebSocket)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
}
