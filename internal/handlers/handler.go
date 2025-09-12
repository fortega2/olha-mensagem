package handlers

import (
	"database/sql"

	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/fortega2/real-time-chat/internal/repository"
)

const (
	reqCtxErrMsg                    = "Request context error"
	reqCtxCancelledOrTimedOutErrMsg = "Request cancelled or timed out"

	failedEncodeuserDataErrMsg     = "Failed to encode user data"
	usernameAndPasswordEmptyErrMsg = "Username and password cannot be empty"
	invalidRequestBodyErrMsg       = "Invalid request body"

	failedEncodeChannelDataErrMsg      = "Failed to encode channel data"
	failedEncodeDeleteChannelRspErrMsg = "Failed to encode delete channel response"

	failedEncodeMessageDataErrMsg = "Failed to encode message data"

	failedEncodeHealthCheckErrMsg = "Failed to encode health check response"
)

type Handler struct {
	logger  logger.Logger
	queries *repository.Queries
	db      *sql.DB
}

func NewHandler(l logger.Logger, q *repository.Queries, db *sql.DB) *Handler {
	return &Handler{
		logger:  l,
		queries: q,
		db:      db,
	}
}
