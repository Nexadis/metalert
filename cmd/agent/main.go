package main

import (
	"context"
	"log"

	"github.com/Nexadis/metalert/internal/agent"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	log.Printf("Build version: %s", buildVersion)
	log.Printf("Build date: %s", buildDate)
	log.Printf("Build commit: %s", buildCommit)
	config := agent.NewConfig()
	config.ParseConfig()
	agent := agent.New(config)
	logger.Info("Agent", config.Address)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log.Fatal(agent.Run(ctx))
}
