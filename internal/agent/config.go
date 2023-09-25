package agent

import (
	"flag"

	"github.com/caarlos0/env/v8"

	"github.com/Nexadis/metalert/internal/utils/logger"
)

// Config содержит в себе конфигурацию агента
type Config struct {
	Address        string `env:"ADDRESS"` // адрес сервера для отправки метрик
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`        // ключ для подписи отправляемых метрик
	RateLimit      int64  `env:"RATE_LIMIT"` // количество воркеров для отправки метрик
}

func NewConfig() *Config {
	return &Config{}
}

// Парсинг командной строки
func (c *Config) parseCmd() {
	flag.StringVar(&c.Address, "a", "localhost:8080", "Server for metrics")
	flag.Int64Var(&c.PollInterval, "p", 2, "Poll Interval")
	flag.Int64Var(&c.ReportInterval, "r", 10, "Report Interval")
	flag.StringVar(&c.Key, "k", "", "Key to sign body")
	flag.Int64Var(&c.RateLimit, "l", 1, "Workers for report")
	flag.Parse()
}

// Парсинг переменных окружения
func (c *Config) parseEnv() {
	err := env.Parse(c)
	if err != nil {
		logger.Error(err.Error())
	}
}

func (c *Config) ParseConfig() {
	c.parseCmd()
	c.parseEnv()
	logger.Info("Parsed Config:",
		"Address", c.Address,
		"ReportInterval", c.ReportInterval,
		"PollInterval", c.PollInterval,
		"Key", c.Key)
}
