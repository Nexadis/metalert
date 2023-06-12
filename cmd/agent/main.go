package main

import (
	"flag"

	"github.com/Nexadis/metalert/internal/agent"
)

var (
	endpoint       string
	pollInterval   int64
	reportInterval int64
)

func parseCmd() {
	flag.StringVar(&endpoint, "a", "http://localhost:8080", "Server for metrics")
	flag.Int64Var(&pollInterval, "p", 2, "Poll Interval")
	flag.Int64Var(&reportInterval, "r", 10, "Report Interval")
	flag.Parse()
}

func main() {
	parseCmd()
	agent := agent.NewAgent(endpoint, pollInterval, reportInterval)
	err := agent.Run()
	if err != nil {
		panic(err)
	}
}
