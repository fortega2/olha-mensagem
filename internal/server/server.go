package server

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/fortega2/real-time-chat/internal/handlers"
	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/fortega2/real-time-chat/internal/repository"
	"github.com/fortega2/real-time-chat/internal/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewServer(l logger.Logger, q *repository.Queries) *Server {
	return &Server{
		logger:  l,
		queries: q,
	}
}

func (s *Server) Start() error {
	r := chi.NewRouter()

	s.configMiddlewares(r)
	s.setRoutes(r)

	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port = ":8080"
	}

	s.server = &http.Server{
		Addr:    port,
		Handler: r,
	}

	s.logger.Info("Starting server on port " + port)

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Shutdown() error {
	if s.server == nil {
		return nil
	}

	s.logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("Failed to gracefully shutdown server. Trying to close the server...", "error", err)
		if closeErr := s.server.Close(); closeErr != nil {
			return closeErr
		}
	}

	s.logger.Info("Server gracefully stopped")
	return nil
}

func (s *Server) configMiddlewares(r *chi.Mux) {
	r.Use(middleware.Recoverer)
}

func (s *Server) setRoutes(r *chi.Mux) {
	handlers := handlers.NewHandler(s.logger, s.queries)
	wsHandler := websocket.NewWebsocketHandler(s.logger, s.queries)

	r.Route("/api", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			r.Post("/login", handlers.LoginUser)
			r.Post("/", handlers.CreateUser)
		})

		r.Route("/channels", func(r chi.Router) {
			r.Get("/", handlers.GetAllChannels)
			r.Post("/", handlers.CreateChannel)
			r.Delete("/{channelId}/users/{userId}", handlers.DeleteChannel)
		})

		r.Route("/messages", func(r chi.Router) {
			r.Get("/history/{channelId}", handlers.GetHistoryMessagesByChannel)
		})

		r.Get("/ws/{channelId}/{userId}", wsHandler.HandleWebSocket)
	})

	r.Mount("/", s.serveStaticFiles())
}
