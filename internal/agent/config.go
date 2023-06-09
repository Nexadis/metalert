package agent

import (
	"flag"

	"github.com/Nexadis/metalert/internal/utils/logger"
	"github.com/caarlos0/env/v8"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) parseCmd() {
	flag.StringVar(&c.Address, "a", "localhost:8080", "Server for metrics")
	flag.Int64Var(&c.PollInterval, "p", 2, "Poll Interval")
	flag.Int64Var(&c.ReportInterval, "r", 10, "Report Interval")
	logger.Info("Parse command flags:",
		"Address", c.Address,
		"ReportInterval", c.ReportInterval,
		"PollInterval", c.PollInterval)
	flag.Parse()
}

func (c *Config) parseEnv() {
	err := env.Parse(c)
	logger.Info("Parse environment:",
		"Address", c.Address,
		"ReportInterval", c.ReportInterval,
		"PollInterval", c.PollInterval)
	if err != nil {
		logger.Error(err.Error())
	}
}

func (c *Config) ParseConfig() {
	c.parseCmd()
	c.parseEnv()
}
