package server_test

import (
	"database/sql"
	"testing"

	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/fortega2/real-time-chat/internal/server"

	_ "github.com/mattn/go-sqlite3"
)

func TestNewServer(t *testing.T) {
	db := initializeTestDB(t)
	defer db.Close()
	mockLogger := logger.NewMockLogger()
	srv := server.NewServer(mockLogger, nil, db)

	if srv == nil {
		t.Fatal("Expected server to be created, got nil")
	}
}

func TestNewServerWithNilLogger(t *testing.T) {
	db := initializeTestDB(t)
	defer db.Close()
	srv := server.NewServer(nil, nil, db)

	if srv == nil {
		t.Fatal("Expected server to be created with nil logger, got nil")
	}
}

func TestNewServerWithNilQueries(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	db := initializeTestDB(t)
	defer db.Close()
	srv := server.NewServer(mockLogger, nil, db)

	if srv == nil {
		t.Fatal("Expected server to be created with nil queries, got nil")
	}
}

func initializeTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	return db
}
