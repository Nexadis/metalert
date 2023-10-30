package storage

import (
	"flag"

	"github.com/caarlos0/env/v8"

	"github.com/Nexadis/metalert/internal/utils/logger"
)

// Config - Конфиг БД
type Config struct {
	StoreInterval   int64  `env:"STORE_INTERVAL" json:"store_interval,omitempty"` // интервал сохранения данных
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"store_file,omitempty"`  // файл для сохранения базы метрик при использовании inmemory хранилища
	Restore         bool   `env:"RESTORE" json:"restore,omitempty"`               // восстановление данных из файл
	DSN             string `env:"DATABASE_DSN" json:"db_dsn,omitempty"`           // Адрес БД
	Retry           int    `env:"DATABASE_CONN_RETRY" json:"db_conn_retries,omitempty"`
	Timeout         int    `env:"DATABASE_TIMEOUT" json:"db_timeout,omitempty"`
}

func NewConfig() *Config {
	return &Config{}
}

var (
	DefaultStoreInterval   = int64(300)
	DefaultFileStoragePath = "/tmp/metrics_db.json"
	DefaultRestore         = true
	DefaultDSN             = ""
	DefaultRetry           = 3
	DefaultTimeout         = 2
)

func (c *Config) ParseCmd(set *flag.FlagSet) {
	set.Int64Var(&c.StoreInterval, "i", DefaultStoreInterval, "Save metrics on disk with interval")
	set.StringVar(&c.FileStoragePath, "f", DefaultFileStoragePath, "File for save metrics")
	set.BoolVar(&c.Restore, "r", DefaultRestore, "Restore file with metrics when start server")
	set.StringVar(&c.DSN, "d", DefaultDSN, "DSN for DB")
	set.IntVar(&c.Retry, "rc", DefaultRetry, "number of repeated attempts to connect to DB")
	set.IntVar(&c.Timeout, "to", DefaultTimeout, "timeout in seconds to connect to DB")
	logger.Info("Parse command flags:",
		"\nStore Interval", c.StoreInterval,
		"\nFile Storage Path", c.FileStoragePath,
		"\nRestore", c.Restore,
		"\nDSN", c.DSN,
	)
}

func (c *Config) ParseEnv() {
	err := env.Parse(c)
	logger.Info("Parse environment:",
		"\nStore Interval", c.StoreInterval,
		"\nFile Storage Path", c.FileStoragePath,
		"\nRestore", c.Restore,
		"\nAddress", c.DSN,
	)
	if err != nil {
		logger.Error(err.Error())
	}
}
