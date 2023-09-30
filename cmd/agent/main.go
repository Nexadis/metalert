package main

import (
	"context"
	"log"

	"github.com/Nexadis/metalert/internal/agent"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

func main() {
	config := agent.NewConfig()
	config.ParseConfig()
	agent := agent.New(config)
	logger.Info("Agent", config.Address)
	log.Fatal(agent.Run(context.Background()))
}
