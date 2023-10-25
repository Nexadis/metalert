package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	exit, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM|syscall.SIGINT|syscall.SIGQUIT)
	defer stop()
	logger.Error(agent.Run(exit))
}
