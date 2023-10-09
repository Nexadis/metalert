package main

import (
	"log"

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
	server, err := server.NewServer(config)
	if err != nil {
		log.Fatal(err)
	}
	server.MountHandlers()
	logger.Info("Server", config.Address)
	err = server.Run()
	if err != nil {
		log.Fatal(err)
	}
}
