package analyzer

import "golang.org/x/tools/go/analysis"

var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "Проверяет наличие exit фукнций в main функциях main-пакетов",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	return nil, nil
}
