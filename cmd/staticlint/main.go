// staticlint подключает все анализаторы к multichecker'у
// Для использования достаточно запустить ./main ./...
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
