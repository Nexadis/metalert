package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Nexadis/metalert/internal/server"
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
	config := server.NewConfig()
	config.ParseConfig()
	server, err := server.New(config)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("Server", config.Address)
	exit, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM|syscall.SIGINT|syscall.SIGQUIT)
	defer stop()
	err = server.Run(exit)
	if err != nil {
		log.Fatal(err)
	}
}
