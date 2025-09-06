package database

import (
	"database/sql"
	"os"

	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type database struct {
	db *sql.DB
}

func NewDatabase(logger logger.Logger) (*database, error) {
	dbName := os.Getenv("DB_NAME")

	db, err := initializeDB(dbName)
	if err != nil {
		return nil, err
	}

	if err := migrateDB(dbName, logger); err != nil {
		return nil, err
	}

	logger.Info("Database initialized successfully")
	return &database{
		db: db,
	}, nil
}

func (d *database) GetDB() *sql.DB {
	return d.db
}

func (d *database) Close() error {
	return d.db.Close()
}

func initializeDB(dbName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbName)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return nil, err
	}

	if _, err := db.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA synchronous = NORMAL;`); err != nil {
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA busy_timeout = 5000;`); err != nil {
		return nil, err
	}

	return db, nil
}

func migrateDB(dbName string, logger logger.Logger) error {
	dbMigrationsPath := os.Getenv("DB_MIGRATIONS_PATH")
	migratePath := "file://" + dbMigrationsPath
	m, err := migrate.New(migratePath, "sqlite3://"+dbName)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	logger.Info("Database migrated successfully")
	return nil
}
