package server

import (
	"flag"

	"github.com/Nexadis/metalert/internal/db"
	"github.com/Nexadis/metalert/internal/utils/logger"
	"github.com/caarlos0/env/v8"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int64  `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DB              *db.Config
}

func NewConfig() *Config {
	db := db.NewConfig()
	return &Config{
		DB: db,
	}
}

func (c *Config) parseCmd() {
	flag.StringVar(&c.Address, "a", "localhost:8080", "Server for metrics")
	flag.Int64Var(&c.StoreInterval, "i", 300, "Save metrics on disk with interval")
	flag.StringVar(&c.FileStoragePath, "f", "/tmp/metrics_db.json", "File for save metrics")
	flag.BoolVar(&c.Restore, "r", true, "Restore file with metrics when start server")
	logger.Info("Parse command flags:",
		"Address", c.Address,
		"Store Interval", c.StoreInterval,
		"File Storage Path", c.FileStoragePath,
		"Restore", c.Restore,
	)
}

func (c *Config) parseEnv() {
	err := env.Parse(c)
	logger.Info("Parse environment:",
		"Address", c.Address,
		"Store Interval", c.StoreInterval,
		"File Storage Path", c.FileStoragePath,
		"Restore", c.Restore,
	)
	if err != nil {
		logger.Error(err.Error())
	}
}

func (c *Config) ParseConfig() {
	c.DB.ParseCmd()
	c.parseCmd()
	flag.Parse()
	c.parseEnv()
	c.DB.ParseEnv()
}
