package server

import (
	"database/sql"
	"net/http"

	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/fortega2/real-time-chat/internal/repository"
)

type Server struct {
	logger  logger.Logger
	queries *repository.Queries
	db      *sql.DB
	server  *http.Server
}
