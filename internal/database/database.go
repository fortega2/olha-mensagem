package database

import (
	"database/sql"
	"os"

	"github.com/fortega2/real-time-chat/internal/logger"

	_ "github.com/mattn/go-sqlite3"
)

type database struct {
	db     *sql.DB
	logger logger.Logger
}

func NewDatabase(logger logger.Logger) (*database, error) {
	db, err := initializeDB()

	if err != nil {
		return nil, err
	}

	logger.Info("Database initialized successfully")
	return &database{
		db:     db,
		logger: logger,
	}, nil
}

func (d *database) GetDB() *sql.DB {
	return d.db
}

func (d *database) Close() error {
	return d.db.Close()
}

func initializeDB() (*sql.DB, error) {
	dbName := os.Getenv("DB_NAME")

	db, err := sql.Open("sqlite3", dbName)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
