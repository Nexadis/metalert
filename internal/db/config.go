package db

import (
	"flag"

	"github.com/Nexadis/metalert/internal/utils/logger"
	"github.com/caarlos0/env/v8"
)

type Config struct {
	DSN string `env:"DATABASE_DSN"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) ParseCmd() {
	flag.StringVar(&c.DSN, "d", "", "DSN for DB")
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
