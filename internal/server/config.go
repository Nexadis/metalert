package server

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v8"

	"github.com/Nexadis/metalert/internal/storage"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

// Config - Конфиг сервера
type Config struct {
	Address   string `env:"ADDRESS"`
	Verbose   bool   `env:"VERBOSE"`    // Включить логгирование
	SignKey   string `env:"KEY"`        // Ключ для подписи всех пакетов
	CryptoKey string `env:"CRYPTO_KEY"` // Приватный ключ для расшифровки метрик
	DB        *storage.Config
}

// NewConfig() Конструктор для конфига
func NewConfig() *Config {
	db := storage.NewConfig()
	return &Config{
		DB: db,
	}
}

func (c *Config) parseCmd(set *flag.FlagSet) {
	set.StringVar(&c.Address, "a", "localhost:8080", "Server for metrics")
	set.BoolVar(&c.Verbose, "v", true, "Verbose logging")
	set.StringVar(&c.SignKey, "k", "", "Key to sign body")
	set.StringVar(&c.CryptoKey, "crypto-key", "", "Path to file with private-key")
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
