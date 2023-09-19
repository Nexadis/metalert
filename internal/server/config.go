package server

import (
	"flag"

	"github.com/Nexadis/metalert/internal/storage/db"
	"github.com/Nexadis/metalert/internal/utils/logger"
	"github.com/caarlos0/env/v8"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int64  `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	Verbose         bool   `env:"VERBOSE"`
	Key             string `env:"KEY"`
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
	flag.BoolVar(&c.Verbose, "v", false, "Verbose logging")
	flag.StringVar(&c.Key, "k", "", "Key to sign body")
}

func (c *Config) parseEnv() {
	err := env.Parse(c)
	if err != nil {
		logger.Error(err.Error())
	}
}

func (c *Config) ParseConfig() {
	c.parseCmd()
	flag.Parse()
	c.parseEnv()
	if c.Verbose {
		logger.Enable()
	}
	c.DB.ParseCmd()
	c.DB.ParseEnv()
	logger.Info("Parse config:",
		"Address", c.Address,
		"Store Interval", c.StoreInterval,
		"File Storage Path", c.FileStoragePath,
		"Restore", c.Restore,
		"Verbose", c.Verbose,
		"Key", c.Key,
	)
}
