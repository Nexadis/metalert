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

func (db *DB) Get(ctx context.Context, mtype, id string) (storage.ObjectGetter, error) {
	row := db.db.QueryRowContext(ctx,
		`SELECT delta, value FROM Metrics WHERE type = $1 AND id = $2`,
		mtype, id,
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
		`SELECT * FROM Metrics`,
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
	logger.Info("Begin transaction for db!")
	_, err = db.db.ExecContext(ctx, "INSERT INTO Metrics (id, type, delta, value)"+
		" VALUES ($1,$2,$3,$4)",
		m.ID,
		m.MType,
		m.Delta,
		m.Value,
	)
	if err == nil {
		db.size += 1
		return nil
	}
	_, err = db.db.ExecContext(ctx, "UPDATE Metrics SET delta=delta + $3, value=$4 WHERE id=$1 AND type=$2 ",
		m.ID,
		m.MType,
		m.Delta,
		m.Value,
	)
	return err
}
