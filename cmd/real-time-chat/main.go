package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/fortega2/real-time-chat/internal/database"
	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/fortega2/real-time-chat/internal/repository"
	"github.com/fortega2/real-time-chat/internal/server"
	"github.com/fortega2/real-time-chat/internal/shutdown"
	"github.com/fortega2/real-time-chat/internal/websocket"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	logger := logger.NewSlogLogger()
	logger.Info("Logger initialized with level " + os.Getenv("LOG_LEVEL"))

	if err != nil {
		logger.Info("No .env file found, proceeding with default environment variables")
	} else {
		logger.Info("Environment variables loaded from .env file")
	}

	db, err := database.NewDatabase(logger)
	if err != nil {
		logger.Fatal("Failed to initialize database", "error", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database", "error", err)
		}
	}()

	queries := repository.New(db.GetDB())
	srv := server.NewServer(logger, queries, db.GetDB())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	awaitShutdownSignal(sigChan, logger, srv)
}

func awaitShutdownSignal(sigChan <-chan os.Signal, logger logger.Logger, srv *server.Server) {
	select {
	case <-shutdown.Wait():
		logger.Info("Shutdown signal received from internal package")
	case sig := <-sigChan:
		logger.Info("Shutdown signal received", "signal", sig)
	}

	performCleanup(logger, srv)
}

func performCleanup(logger logger.Logger, srv *server.Server) {
	logger.Info("Starting cleanup process...")

	websocket.Shutdown()

	if err := srv.Shutdown(); err != nil {
		logger.Error("Failed to shutdown server", "error", err)
	}

	logger.Info("Cleanup completed")
}
