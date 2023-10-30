package main

import (
	"flag"
	"log"

	"github.com/Nexadis/metalert/internal/utils/asymcrypt"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

func main() {
	var keyfile string
	logger.Enable()
	flag.StringVar(&keyfile, "o", "key", "Prefix for public and private keys")
	flag.Parse()
	log.Fatal(asymcrypt.NewPem(keyfile))
}
