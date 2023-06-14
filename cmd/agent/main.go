package main

import "github.com/Nexadis/metalert/internal/agent"

func main() {
	agent := agent.NewAgent("http://localhost:8080", 2, 10)
	err := agent.Run()
	if err != nil {
		panic(err)
	}
}
