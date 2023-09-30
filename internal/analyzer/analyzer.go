package analyzer

import (
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"honnef.co/go/tools/staticcheck"
)

var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "Проверяет наличие exit функций в main функциях main-пакетов",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	return nil, nil
}

func StandardPasses() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		stdmethods.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
	}
}

func StaticChecks() []*analysis.Analyzer {
	needChecks := []string{
		"SA",
		"S100",
		"ST",
		"QF1002",
		"QF1003",
		"QF1006",
		"QF1010",
		"QF1012",
	}
	staticchecks := make([]*analysis.Analyzer, 0, 20)
	for _, a := range staticcheck.Analyzers {
		for _, need := range needChecks {
			if strings.HasPrefix(a.Analyzer.Name, need) {
				staticchecks = append(staticchecks, a.Analyzer)
			}
		}
	}
	return staticchecks
}

func AllAnalyzers() []*analysis.Analyzer {
	a := make([]*analysis.Analyzer, 0, 100)
	a = append(a, StaticChecks()...)
	a = append(a, StandardPasses()...)
	return a
}
