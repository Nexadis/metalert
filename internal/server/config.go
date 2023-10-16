package server

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v8"

	"github.com/Nexadis/metalert/internal/storage/db"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

// Config - Конфиг сервера
type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int64  `env:"STORE_INTERVAL"`    // интервал сохранения данных
	FileStoragePath string `env:"FILE_STORAGE_PATH"` // файл для сохранения базы метрик при использовании inmemory хранилища
	Restore         bool   `env:"RESTORE"`           // восстановление данных из файл
	Verbose         bool   `env:"VERBOSE"`           // Включить логгирование
	SignKey         string `env:"KEY"`               // Ключ для подписи всех пакетов
	CryptoKey       string `env:"CRYPTO_KEY"`        // Приватный ключ для расшифровки метрик
	DB              *db.Config
}

// NewConfig() Конструктор для конфига
func NewConfig() *Config {
	db := db.NewConfig()
	return &Config{
		DB: db,
	}
}

func (c *Config) parseCmd(set *flag.FlagSet) {
	set.StringVar(&c.Address, "a", "localhost:8080", "Server for metrics")
	set.Int64Var(&c.StoreInterval, "i", 300, "Save metrics on disk with interval")
	set.StringVar(&c.FileStoragePath, "f", "/tmp/metrics_db.json", "File for save metrics")
	set.BoolVar(&c.Restore, "r", true, "Restore file with metrics when start server")
	set.BoolVar(&c.Verbose, "v", true, "Verbose logging")
	set.StringVar(&c.SignKey, "k", "", "Key to sign body")
	flag.StringVar(&c.CryptoKey, "crypto-key", "", "Path to file with public-key")
}

func (c *Config) parseEnv() {
	err := env.Parse(c)
	if err != nil {
		logger.Error(err.Error())
	}
}

// ParseConfig() выполняет парсинг всех конфига сервера
func (c *Config) ParseConfig() {
	set := c.setFlags()
	set.Parse(os.Args[1:])
	c.parseEnv()
	c.DB.ParseEnv()
	if c.Verbose {
		logger.Enable()
	}
	logger.Info("Parse config:",
		"\nAddress", c.Address,
		"\nStore Interval", c.StoreInterval,
		"\nFile Storage Path", c.FileStoragePath,
		"\nRestore", c.Restore,
		"\nVerbose", c.Verbose,
		"\nSign Key", c.SignKey,
		"\nCrypto Key", c.CryptoKey,
	)
}

func (c *Config) setFlags() *flag.FlagSet {
	set := flag.NewFlagSet("", flag.ExitOnError)
	c.parseCmd(set)
	c.DB.ParseCmd(set)
	return set
}

func (c *Config) SetDefault() {
	set := c.setFlags()
	set.Parse([]string{})
	if c.Verbose {
		logger.Enable()
	}
}
