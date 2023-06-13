package main

import (
	"github.com/Nexadis/metalert/internal/agent"
	"github.com/Nexadis/metalert/internal/utils/config"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

func main() {
	config.ParseConfig()
	agent := agent.NewAgent(
		config.MainConfig.Address,
		config.MainConfig.PollInterval,
		config.MainConfig.ReportInterval)
	logger.Info("Start agent on %s", config.MainConfig.Address)
	err := agent.Run()
	if err != nil {
		panic(err)
	}
}
