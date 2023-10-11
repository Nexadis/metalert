// Пакет для работы с PostgreSQL
package db

import (
	"flag"

	"github.com/caarlos0/env/v8"

	"github.com/Nexadis/metalert/internal/utils/logger"
)

// Config - Конфиг БД
type Config struct {
	DSN     string `env:"DATABASE_DSN"` // Адрес БД
	Retry   int    `env:"DATABASE_CONN_RETRY"`
	Timeout int    `env:"DATABASE_TIMEOUT"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) ParseCmd(set *flag.FlagSet) {
	set.StringVar(&c.DSN, "d", "", "DSN for DB")
	set.IntVar(&c.Retry, "rc", 3, "number of repeated attempts to connect to DB")
	set.IntVar(&c.Timeout, "to", 2, "timeout in seconds to connect to DB")
	logger.Info("Parse command flags:",
		"DSN", c.DSN,
	)
}

func (c *Config) ParseEnv() {
	err := env.Parse(c)
	logger.Info("Parse environment:",
		"Address", c.DSN,
	)
	if err != nil {
		logger.Error(err.Error())
	}
}
