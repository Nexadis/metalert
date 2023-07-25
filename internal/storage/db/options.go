package db

import "time"

func SetTimeout(timeout time.Duration) func(*DB) {
	return func(db *DB) {
		db.conn.timeout = timeout
	}
}

func SetRetries(retries int) func(*DB) {
	return func(db *DB) {
		db.conn.retries = retries
	}
}
