package main

import (
	"github.com/Nexadis/metalert/internal/agent"
	"github.com/Nexadis/metalert/internal/utils/config"
)

func main() {
	config.ParseConfig()
	agent := agent.NewAgent(
		config.MainConfig.Address,
		config.MainConfig.PollInterval,
		config.MainConfig.ReportInterval)
	err := agent.Run()
	if err != nil {
		panic(err)
	}
}
