package server

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/caarlos0/env/v8"

	"github.com/Nexadis/metalert/internal/storage"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

// Config - Конфиг сервера
type Config struct {
	Address   string          `env:"ADDRESS" json:"address,omitempty"`
	Verbose   bool            `env:"VERBOSE" json:"verbose,omitempty"`       // Включить логгирование
	SignKey   string          `env:"KEY" json:"key,omitempty"`               // Ключ для подписи всех пакетов
	CryptoKey string          `env:"CRYPTO_KEY" json:"crypto_key,omitempty"` // Приватный ключ для расшифровки метрик
	Config    string          `env:"CONFIG"`
	DB        *storage.Config `json:"db,omitempty"`
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
	set.StringVar(&c.Config, "config", "", "Path to file with config")
}

func (c *Config) parseEnv() {
	err := env.Parse(c)
	if err != nil {
		logger.Error(err.Error())
	}
}

func (c *Config) parseFile() {
	set := flag.NewFlagSet("", flag.ContinueOnError)
	set.StringVar(&c.Config, "config", "", "Path to file with config")
	set.Parse(os.Args[1:])
	c.loadJSON()
}

// ParseConfig() выполняет парсинг всех конфигов сервера
func (c *Config) ParseConfig() {
	c.parseFile()
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

func (c *Config) loadJSON() {
	if c.Config == "" {
		return
	}
	data, err := os.ReadFile(c.Config)
	if err != nil {
		logger.Error(err)
		return
	}
	err = json.Unmarshal(data, c)
	if err != nil {
		logger.Error(err)
	}
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
