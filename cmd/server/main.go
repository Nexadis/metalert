package main

import "github.com/Nexadis/metalert/internal/server"

func main() {
	server := server.NewServer(":8080")
	server.MountHandlers()
	err := server.Run()
	if err != nil {
		panic(err)
	}

}
