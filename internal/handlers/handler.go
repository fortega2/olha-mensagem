package handlers

import (
	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/fortega2/real-time-chat/internal/repository"
)

type Handler struct {
	logger  logger.Logger
	queries *repository.Queries
}

func NewHandler(l logger.Logger, q *repository.Queries) *Handler {
	return &Handler{
		logger:  l,
		queries: q,
	}
}
