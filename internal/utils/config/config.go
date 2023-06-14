package config

import (
	"flag"

	"github.com/caarlos0/env/v8"

	"github.com/Nexadis/metalert/internal/utils/logger"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
}

var MainConfig = Config{}

func parseCmd() {
	flag.StringVar(&MainConfig.Address, "a", "localhost:8080", "Server for metrics")
	flag.Int64Var(&MainConfig.PollInterval, "p", 2, "Poll Interval")
	flag.Int64Var(&MainConfig.ReportInterval, "r", 10, "Report Interval")
	logger.Debug("Get values from flags:\n%s=%s\n%s=%d\n%s=%d\n",
		"Address", MainConfig.Address,
		"ReportInterval", MainConfig.ReportInterval,
		"PollInterval", MainConfig.PollInterval)
	flag.Parse()
}

func parseEnv() {
	err := env.Parse(&MainConfig)
	logger.Debug("Get values from env:\n%s=%s\n%s=%d\n%s=%d\n",
		"Address", MainConfig.Address,
		"ReportInterval", MainConfig.ReportInterval,
		"PollInterval", MainConfig.PollInterval)
	if err != nil {
		logger.Error(err.Error())
	}
}

func ParseConfig() {
	parseCmd()
	parseEnv()
}
