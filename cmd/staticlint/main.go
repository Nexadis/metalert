package main

import (
	"github.com/Nexadis/metalert/internal/analyzer"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(
		analyzer.AllAnalyzers()...,
	)
}
