package db

import "time"

// SetTimeout Устанавливает таймаут для БД.
func SetTimeout(timeout time.Duration) func(*DB) {
	return func(db *DB) {
		db.conn.timeout = timeout
	}
}

// SetRetries Устанавливает количество повторных попыток для БД.
func SetRetries(retries int) func(*DB) {
	return func(db *DB) {
		db.conn.retries = retries
	}
}
