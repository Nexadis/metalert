package main

import "github.com/Nexadis/metalert/internal/agent"

func main() {
	parseCmd()
	agent := agent.NewAgent(endpoint, pollInterval, reportInterval)
	err := agent.Run()
	if err != nil {
		panic(err)
	}
}
