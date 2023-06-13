package main

import (
	"github.com/Nexadis/metalert/internal/server"
	"github.com/Nexadis/metalert/internal/utils/config"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

func main() {
	config.ParseConfig()
	server := server.NewServer(config.MainConfig.Address)
	server.MountHandlers()
	logger.Info("Start server on %s", config.MainConfig.Address)
	err := server.Run()
	if err != nil {
		panic(err)
	}

}
