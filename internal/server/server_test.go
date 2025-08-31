package server_test

import (
	"testing"

	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/fortega2/real-time-chat/internal/server"
)

func TestNewServer(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	srv := server.NewServer(mockLogger, nil)

	if srv == nil {
		t.Fatal("Expected server to be created, got nil")
	}
}

func TestNewServerWithNilLogger(t *testing.T) {
	srv := server.NewServer(nil, nil)

	if srv == nil {
		t.Fatal("Expected server to be created with nil logger, got nil")
	}
}

func TestNewServerWithNilQueries(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	srv := server.NewServer(mockLogger, nil)

	if srv == nil {
		t.Fatal("Expected server to be created with nil queries, got nil")
	}
}
