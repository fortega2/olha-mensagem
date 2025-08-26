package server

import (
	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/fortega2/real-time-chat/internal/repository"
)

type Server struct {
	logger  logger.Logger
	queries *repository.Queries
}
