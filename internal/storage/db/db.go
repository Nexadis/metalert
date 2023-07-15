package db

import (
	"context"
	"database/sql"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/storage"
	"github.com/Nexadis/metalert/internal/utils/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const schema = `CREATE TABLE Metrics(
"id" VARCHAR(250) NOT NULL,
"type" VARCHAR(100) NOT NULL,
"delta" DOUBLE PRECISION,
"value" BIGINT,
CONSTRAINT ID PRIMARY KEY (id,type));
`

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
		db:   db,
		size: 0,
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

func (db *DB) Get(ctx context.Context, mtype, id string) (storage.ObjectGetter, error) {
	stmt, err := db.db.PrepareContext(ctx,
		`SELECT delta, value FROM Metrics WHERE type = $1 AND id = $2`)
	if err != nil {
		return nil, err
	}
	row := stmt.QueryRowContext(ctx,
		mtype, id,
	)
	m := &metrx.Metrics{}
	err = row.Scan(&m.Delta, &m.Value)
	if err != nil {
		return nil, err
	}
	return m.GetMetricsString()
}

func (db *DB) GetAll(ctx context.Context) ([]storage.ObjectGetter, error) {
	stmt, err := db.db.PrepareContext(ctx,
		`SELECT * FROM Metrics`)
	if err != nil {
		return nil, err
	}
	metrics := make([]storage.ObjectGetter, 0, db.size)
	rows, err := stmt.QueryContext(ctx)
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
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (db *DB) Set(ctx context.Context, mtype, id, value string) error {
	m := metrx.Metrics{}
	err := m.ParseMetricsString(
		&metrx.MetricsString{
			ID:    id,
			MType: mtype,
			Value: value,
		},
	)
	if err != nil {
		return err
	}
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO Metrics (id, type, delta, value) "+
		"VALUES ($1,$2,$3,$4) ON CONFLICT(id,type) "+
		"DO UPDATE SET delta=metrics.delta + $3, value=$4",
	)
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx,
		m.ID,
		m.MType,
		m.Delta,
		m.Value,
	)
	if err != nil {
		return err
	}
	db.size += 1
	return tx.Commit()
}
