package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/Nexadis/metalert/internal/agent/client"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("Pass address of grpc server in arg", flag.Args())
	}
	addr := flag.Args()[0]
	gc := client.NewGRPC(addr)
	ms, err := gc.Get(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	logger.Info(fmt.Printf("GRPC metrics:\n%v", ms))
}
