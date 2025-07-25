package db

import (
	"database/sql"
	"log/slog"

	_ "github.com/lib/pq"
)

type Postgres struct {
	Db *sql.DB
}

func Connect(PostgresUri string) (*Postgres, error) {

	db, err := sql.Open("postgres", PostgresUri)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	slog.Info("Connected to the database.")

	return &Postgres{Db: db}, nil
}