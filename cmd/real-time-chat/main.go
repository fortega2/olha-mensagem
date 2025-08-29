package main

import (
	"os"

	"github.com/fortega2/real-time-chat/internal/database"
	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/fortega2/real-time-chat/internal/repository"
	"github.com/fortega2/real-time-chat/internal/server"
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
	srv := server.NewServer(logger, queries)

	if err := srv.Start(); err != nil {
		logger.Fatal("Failed to start server", "error", err)
	}
}
