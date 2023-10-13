package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/storage"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

// schema - Схема для метрик
const schema = `CREATE TABLE Metrics(
"id" VARCHAR(250) NOT NULL,
"type" VARCHAR(100) NOT NULL,
"delta" BIGINT,
"value" DOUBLE PRECISION,
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

// DB Реализует логику работы с БД.
type DB struct {
	db   *sql.DB
	size int
	conn connection
}

var _ DataBase = &DB{}

type connection struct {
	retries int
	timeout time.Duration
}

// New Конструктор для БД. Настраивает политики retry
func New(config *Config) *DB {
	db := &sql.DB{}
	DB := &DB{
		db:   db,
		size: 0,
	}
	Configure(DB,
		SetRetries(config.Retry),
		SetTimeout(time.Duration(config.Timeout)),
	)
	return DB
}

func Configure(db *DB, options ...func(*DB)) {
	for _, o := range options {
		o(db)
	}
}

// Open Открывает подключение к БД. Создаёт схемы
func (db *DB) Open(ctx context.Context, DSN string) error {
	var pgx *sql.DB
	err := db.retry(func() error {
		var err error
		pgx, err = sql.Open("pgx", DSN)
		if err != nil {
			logger.Info("Unable to connect to database:", err)
			return err
		}
		return pgx.Ping()
	})
	logger.Info("Connect to:", DSN)
	if err != nil {
		logger.Error("Unable to connect to database:", err)
		return err
	}
	db.db = pgx
	_, err = pgx.ExecContext(ctx, schema)
	if err != nil {
		logger.Error("Unable to create table:", err)
	}
	return nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) Ping() error {
	return db.retry(db.db.Ping)
}

// Get Получает значение метрики из БД.
func (db *DB) Get(ctx context.Context, mtype, id string) (metrx.Metric, error) {
	stmt, err := db.db.PrepareContext(ctx,
		`SELECT delta, value FROM Metrics WHERE type=$1 AND id= $2`)
	if err != nil {
		return metrx.Metric{}, err
	}
	var row *sql.Row
	err = db.retry(func() error {
		row = stmt.QueryRowContext(ctx,
			mtype, id,
		)
		err = row.Err()
		if err != nil {
			return checkConnection(err)
		}
		return nil
	})
	if err != nil {
		return metrx.Metric{}, err
	}
	m := metrx.Metric{
		ID:    id,
		MType: mtype,
	}
	err = row.Scan(&m.Delta, &m.Value)
	if err != nil {
		return metrx.Metric{}, err
	}
	return m, nil
}

// GetAll Получает все метрики из БД.
func (db *DB) GetAll(ctx context.Context) ([]metrx.Metric, error) {
	stmt, err := db.db.PrepareContext(ctx,
		`SELECT * FROM Metrics`)
	if err != nil {
		return nil, err
	}
	metrics := make([]metrx.Metric, 0, db.size)
	var rows *sql.Rows
	err = db.retry(func() error {
		rows, err = stmt.QueryContext(ctx)
		if err != nil {
			return checkConnection(err)
		}
		if rows.Err() != nil {
			return checkConnection(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		metric := metrx.Metric{}
		err = rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
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

// Set Обновляет метрику в БД.
func (db *DB) Set(ctx context.Context, m metrx.Metric) error {
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
	err = db.retry(func() error {
		_, err = stmt.ExecContext(ctx,
			m.ID,
			m.MType,
			m.Delta,
			m.Value,
		)
		if err != nil {
			return checkConnection(err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	db.size += 1
	return tx.Commit()
}

func checkConnection(err error) error {
	if pgerrcode.IsConnectionException(err.Error()) {
		return fmt.Errorf("db conenction problem %w", err)
	}
	return err
}

// retry Выполняет функцию, обернув её в retry
func (db *DB) retry(fn func() error) error {
	attempt := db.conn.retries
	var err error
	for attempt > 0 {
		err = fn()
		if err != nil {
			logger.Info("Retry", attempt)
			attempt--
			time.Sleep(db.conn.timeout)
		} else {
			logger.Info("Don't retry all is good ")
			return nil
		}
	}
	return fmt.Errorf("retry db %w", err)
}
