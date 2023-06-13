package main

import (
	"github.com/Nexadis/metalert/internal/server"
	"github.com/Nexadis/metalert/internal/utils/config"
)

func main() {
	config.ParseConfig()
	server := server.NewServer(config.MainConfig.Address)
	server.MountHandlers()
	err := server.Run()
	if err != nil {
		panic(err)
	}

}
