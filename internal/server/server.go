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
	"github.com/fortega2/real-time-chat/internal/repository"
	"github.com/fortega2/real-time-chat/internal/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	logger  logger.Logger
	queries *repository.Queries
}

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

func (s *Server) configMiddlewares(r *chi.Mux) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
}

func (s *Server) setRoutes(r *chi.Mux) {
	handlers := handlers.NewHandler(s.logger, s.queries)

	r.Get("/", handlers.Root)

	r.Post("/login", handlers.LoginUser)
	r.Get("/users/{id}", handlers.GetUserByID)
	r.Post("/users", handlers.CreateUser)

	r.Get("/ws/{userId}", websocket.HandleWebSocket)
}
