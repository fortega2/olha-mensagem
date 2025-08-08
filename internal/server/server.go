package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fortega2/real-time-chat/internal/handlers"
	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/fortega2/real-time-chat/internal/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type server struct {
	logger logger.Logger
}

func NewServer(l logger.Logger) *server {
	return &server{
		logger: l,
	}
}

func (s *server) Start() error {
	r := chi.NewRouter()

	s.configMiddlewares(r)
	s.setRoutes(r)

	port := ":8080"

	server := &http.Server{
		Addr:    port,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s.logger.Info("Starting server on port " + port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to start listening the server", "error", err)
		}
	}()

	<-quit

	websocket.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.logger.Info("Shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		s.logger.Error("Failed to gracefully shutdown server. Trying to close the server...", "error", err)
		if closeErr := server.Close(); closeErr != nil {
			s.logger.Fatal("Failed to close server", "error", closeErr)
		}
	}

	s.logger.Info("Server gracefully stopped")
	return nil
}

func (s *server) configMiddlewares(r *chi.Mux) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
}

func (s *server) setRoutes(r *chi.Mux) {
	handlers := handlers.NewHandler(s.logger)

	r.Get("/", handlers.Root)
	r.Post("/login", handlers.Login)
	r.Get("/users/{id}", handlers.GetUserByID)
	r.Get("/ws/{userId}", websocket.HandleWebSocket)
}
