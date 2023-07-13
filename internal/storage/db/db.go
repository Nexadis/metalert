package db

import (
	"context"
	"database/sql"

	"github.com/Nexadis/metalert/internal/storage"
	"github.com/Nexadis/metalert/internal/utils/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBOpener interface {
	Open(ctx context.Context, DSN string) error
}

type DBCloser interface {
	Close() error
}

type DBPing interface {
	Ping() error
}

type DataBase interface {
	DBOpener
	DBPing
	storage.Storage
	DBCloser
}

type DB struct {
	db *sql.DB
}

func New() DataBase {
	db := &sql.DB{}
	return &DB{
		db: db,
	}
}

func (db *DB) Open(ctx context.Context, DSN string) error {
	pgx, err := sql.Open("pgx", DSN)
	logger.Info("Connect to:", DSN)
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

func (db *DB) Ping() error {
	return db.db.Ping()
}

func (db *DB) Get(ctx context.Context, valType, name string) (storage.ObjectGetter, error) {
	return nil, nil
}

func (db *DB) GetAll(ctx context.Context) ([]storage.ObjectGetter, error) {
	return nil, nil
}

func (db *DB) Set(ctx context.Context, vlaType, name, value string) error {
	return nil
}
