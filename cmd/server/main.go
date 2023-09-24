package main

import (
	"github.com/Nexadis/metalert/internal/server"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

func main() {
	config := server.NewConfig()
	config.ParseConfig()
	server, err := server.NewServer(config)
	if err != nil {
		panic(err)
	}
	server.MountHandlers()
	logger.Info("Server", config.Address)
	err = server.Run()
	if err != nil {
		panic(err)
	}
}
