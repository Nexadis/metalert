package db

import (
	"flag"

	"github.com/caarlos0/env/v8"

	"github.com/Nexadis/metalert/internal/utils/logger"
)

type Config struct {
	DSN     string `env:"DATABASE_DSN"`
	Retry   int    `env:"DATABASE_CONN_RETRY"`
	Timeout int    `env:"DATABASE_TIMEOUT"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) ParseCmd() {
	flag.StringVar(&c.DSN, "d", "", "DSN for DB")
	flag.IntVar(&c.Retry, "rc", 3, "number of repeated attempts to connect to DB")
	flag.IntVar(&c.Timeout, "to", 2, "timeout in seconds to connect to DB")
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

func (c *Config) ParseConfig() {
	c.ParseCmd()
	c.ParseEnv()
}
