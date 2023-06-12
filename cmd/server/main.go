package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Nexadis/metalert/internal/server"
	"github.com/caarlos0/env/v8"
)

type Config struct {
	Address string `env:"ADDRESS"`
}

var MainConfig = Config{}

func parseCmd() {
	flag.StringVar(&MainConfig.Address, "a", "localhost:8080", "Server for metrics")
	fmt.Printf("Get values from flags:\n%s=%s",
		"Address", MainConfig.Address,
	)
	flag.Parse()
}

func parseEnv() {
	err := env.Parse(&MainConfig)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Get values from env:\n%s=%s",
		"Address", MainConfig.Address,
	)
}
func parseConfig() {
	parseCmd()
	parseEnv()
}

func main() {
	parseConfig()
	server := server.NewServer(MainConfig.Address)
	server.MountHandlers()
	err := server.Run()
	if err != nil {
		panic(err)
	}

}
