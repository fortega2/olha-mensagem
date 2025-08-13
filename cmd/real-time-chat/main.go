package main

import (
	"github.com/fortega2/real-time-chat/internal/database"
	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/fortega2/real-time-chat/internal/repository"
	"github.com/fortega2/real-time-chat/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	logger := logger.NewSlogLogger()

	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found, proceeding with default environment variables", "error", err)
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
	srv := server.NewServer(logger, queries)

	if err := srv.Start(); err != nil {
		logger.Fatal("Failed to start server", "error", err)
	}
}
