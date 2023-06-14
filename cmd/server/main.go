package main

import (
	"flag"

	"github.com/Nexadis/metalert/internal/server"
)

var (
	endpoint string
)

func parseCmd() {

	flag.StringVar(&endpoint, "a", "localhost:8080", "Endpoint address")
	flag.Parse()
}
func main() {
	parseCmd()
	server := server.NewServer(endpoint)
	server.MountHandlers()
	err := server.Run()
	if err != nil {
		panic(err)
	}

}
