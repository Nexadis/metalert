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

var (
	defaultAddress   = "localhost:8080"
	defaultVerbose   = true
	defaultSignKey   = ""
	defaultCryptoKey = ""
	defaultConfig    = ""
)

func (c *Config) parseCmd(set *flag.FlagSet) {
	set.StringVar(&c.Address, "a", defaultAddress, "Server for metrics")
	set.BoolVar(&c.Verbose, "v", defaultVerbose, "Verbose logging")
	set.StringVar(&c.SignKey, "k", defaultSignKey, "Key to sign body")
	set.StringVar(&c.CryptoKey, "crypto-key", defaultCryptoKey, "Path to file with private-key")
	set.StringVar(&c.Config, "config", defaultConfig, "Path to file with config")
}

func (c *Config) parseEnv() {
	err := env.Parse(c)
	if err != nil {
		logger.Error(err.Error())
	}
}

func (c *Config) parseFile(set *flag.FlagSet) {
	c.loadJSON()
}

func (c *Config) copyJSON() {
	tmp := NewConfig()
	tmp.Config = c.Config
	tmp.loadJSON()
	if tmp.Address != "" {
		if c.Address == defaultAddress {
			c.Address = tmp.Address
		}
	}
	if c.SignKey == defaultSignKey {
		c.SignKey = tmp.SignKey
	}
	if tmp.Address != "" {
		if c.Address == defaultAddress {
			c.Address = tmp.Address
		}
	}
	if tmp.Address != "" {
		if c.Address == defaultAddress {
			c.Address = tmp.Address
		}
	}
}

// ParseConfig() выполняет парсинг всех конфигов сервера
func (c *Config) ParseConfig() {
	set := c.setFlags()
	set.Parse(os.Args[1:])
	c.parseFile(set)
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
