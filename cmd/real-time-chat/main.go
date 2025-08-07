package main

import (
	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/fortega2/real-time-chat/internal/server"
)

func main() {
	logger := logger.NewSlogLogger()
	srv := server.NewServer(logger)

	if err := srv.Start(); err != nil {
		logger.Fatal("Failed to start server", "error", err)
	}
}
