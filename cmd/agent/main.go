package main

import (
	"context"

	"github.com/Nexadis/metalert/internal/agent"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

func main() {
	config := agent.NewConfig()
	config.ParseConfig()
	agent := agent.New(config)
	logger.Info("Agent", config.Address)
	err := agent.Run(context.Background())
	if err != nil {
		panic(err)
	}
}
