package db

import "time"

// Устанавливает таймаут для БД.
func SetTimeout(timeout time.Duration) func(*DB) {
	return func(db *DB) {
		db.conn.timeout = timeout
	}
}

// Устанавливает количество повторных попыток для БД.
func SetRetries(retries int) func(*DB) {
	return func(db *DB) {
		db.conn.retries = retries
	}
}
