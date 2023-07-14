package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/storage"
	"github.com/Nexadis/metalert/internal/utils/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const table_name = "Metrics"

var schema = fmt.Sprintf(`CREATE TABLE %s(
"name" VARCHAR(250) NOT NULL,
"type" VARCHAR(100) NOT NULL,
"delta" DOUBLE PRECISION,
"value" BIGINT,
CONSTRAINT ID PRIMARY KEY (name,type));
`, table_name)

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
	db   *sql.DB
	size int
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
		logger.Error("Unable to connect to database:", err)
		return err
	}
	db.db = pgx
	_, err = pgx.ExecContext(ctx, schema)
	if err != nil {
		logger.Error("Unable to create table:", err)
		return err
	}
	return nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) Ping() error {
	return db.db.Ping()
}

func (db *DB) Get(ctx context.Context, valType, name string) (storage.ObjectGetter, error) {
	row := db.db.QueryRowContext(ctx,
		`SELECT delta, value FROM $1 WHERE type = $2 AND name = $3`,
		table_name, valType, name,
	)
	m := &metrx.Metrics{}
	err := row.Scan(&m.Delta, &m.Value)
	if err != nil {
		return nil, err
	}
	return m.GetMetricsString()
}

func (db *DB) GetAll(ctx context.Context) ([]storage.ObjectGetter, error) {
	metrics := make([]storage.ObjectGetter, 0, db.size)
	rows, err := db.db.QueryContext(ctx,
		`SELECT * FROM $1`, table_name,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		m := &metrx.Metrics{}
		err = rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value)
		if err != nil {
			return nil, err
		}
		metric, err := m.GetMetricsString()
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	return metrics, nil
}

func (db *DB) Set(ctx context.Context, valType, name, value string) error {
	t, err := db.db.Begin()
	if err != nil {
		return err
	}
	t.ExecContext(ctx, "INSERT INTO ")
	db.size += 1
	return nil
}
