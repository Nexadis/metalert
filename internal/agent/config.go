package agent

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v8"

	"github.com/Nexadis/metalert/internal/utils/logger"
)

// Config содержит в себе конфигурацию агента
type Config struct {
	Address        string        `env:"ADDRESS"` // адрес сервера для отправки метрик
	ReportInterval int64         `env:"REPORT_INTERVAL"`
	PollInterval   int64         `env:"POLL_INTERVAL"`
	Key            string        `env:"KEY"`        // ключ для подписи отправляемых метрик
	CryptoKey      string        `env:"CRYPTO_KEY"` // ключ для шифрования трафика
	RateLimit      int64         `env:"RATE_LIMIT"` // количество воркеров для отправки метрик
	Verbose        bool          `env:"VERBOSE"`    // Включить логгирование
	Transport      TransportType `env:"TRANSPORT"`  // тип транспорта для передачи метрик
}

func NewConfig() *Config {
	return &Config{
		Transport: JSONType,
	}
}

// parseCmd парсит командную строку
func (c *Config) parseCmd() {
	flag.StringVar(&c.Address, "a", "localhost:5533", "Server for metrics (default GRPC address)")
	flag.Int64Var(&c.PollInterval, "p", 2, "Poll Interval")
	flag.Int64Var(&c.ReportInterval, "r", 10, "Report Interval")
	flag.StringVar(&c.Key, "k", "", "Key to sign body")
	flag.StringVar(&c.CryptoKey, "crypto-key", "", "Path to file with public-key")
	flag.Int64Var(&c.RateLimit, "l", 1, "Workers for report")
	flag.BoolVar(&c.Verbose, "v", true, "Verbose logging")
	flag.Var(&c.Transport, "t", fmt.Sprintf("Choose type of transport for posting metrics: %v", Transports))
	flag.Parse()
}

// parseEnv парсит переменные окружения
func (c *Config) parseEnv() {
	err := env.Parse(c)
	if err != nil {
		logger.Error(err.Error())
	}
}

func (c *Config) ParseConfig() {
	c.parseCmd()
	c.parseEnv()
	if c.Verbose {
		logger.Enable()
	}
	logger.Info("Parsed Config:",
		"\nAddress", c.Address,
		"\nReportInterval", c.ReportInterval,
		"\nPollInterval", c.PollInterval,
		"\nKey", c.Key,
		"\nTransport", c.Transport,
	)
}
