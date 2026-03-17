package db

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
)

type Postgres struct {
	Db *sql.DB
}

func Connect(PostgresUri string) (*Postgres, error) {
	db, err := sql.Open("postgres", PostgresUri)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool for optimal performance
	config := DefaultPoolConfig()
	ConfigurePool(db, config)

	slog.Info("Connected to the database", 
		slog.Int("max_open_conns", config.MaxOpenConns),
		slog.Int("max_idle_conns", config.MaxIdleConns))

	return &Postgres{Db: db}, nil
}