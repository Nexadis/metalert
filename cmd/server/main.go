package main

import (
	"github.com/Nexadis/metalert/internal/server"
)

func main() {
	parseCmd()
	server := server.NewServer(endpoint)
	server.MountHandlers()
	err := server.Run()
	if err != nil {
		panic(err)
	}

}
