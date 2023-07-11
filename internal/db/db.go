package db

import (
	"context"
	"database/sql"

	"github.com/Nexadis/metalert/internal/utils/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBOpener interface {
	Open(ctx context.Context, DSN string) error
}

type DBCloser interface {
	Close() error
}

type DataBase interface {
	DBOpener
	DBCloser
}

type DB struct {
	db *sql.DB
}

func NewDB() DataBase {
	db := &sql.DB{}
	return &DB{
		db: db,
	}
}

func (db *DB) Open(ctx context.Context, DSN string) error {
	pgx, err := sql.Open("pgx", DSN)
	if err != nil {
		logger.Error("Unable to connect to database: %v\n", err)
		return err
	}
	db.db = pgx
	return nil
}

func (db *DB) Close() error {
	return db.db.Close()
}
