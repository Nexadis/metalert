// analyzer содержит реализацию собственных анализаторов и подключает все используемые чекеры.
package analyzer

import (
	"go/ast"
	"strings"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
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

// ExitCheckAnalyzer проверяет налачие os.Exit() в main функциях main пакетов.
var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "Проверяет наличие os.Exit() функций в main функциях main пакетов",
	Run:  run,
}

// Analyzers содержит все Разработанные анализаторы.
var Analyzers = []*analysis.Analyzer{}

func init() {
	Analyzers = []*analysis.Analyzer{
		ExitCheckAnalyzer,
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if !isMainPackage(file) {
			continue
		}
		inMain := false
		ast.Inspect(file, func(node ast.Node) bool {
			if isNameFunc(node, "main") {
				inMain = true
				return true
			}
			if inMain && isNameFunc(node, "Exit") {
				pass.Reportf(node.Pos(), "don't use exit in main")
			}

			return true
		})
	}
	return nil, nil
}

func isMainPackage(file *ast.File) bool {
	return file.Name.Name == "main"
}

func isNameFunc(node ast.Node, name string) bool {
	if id, ok := node.(*ast.Ident); ok {
		if id.Name == name {
			return true
		}
	}
	return false
}

// StandardPasses возвращает все анализаторы из стандартной библиотеки.
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

// StaticChecks Возвращает используемые анализаторы из staticheck.io.
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

// ThirdChecks возвращает сторонние подключаемые анализаторы.
func ThirdChecks() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		bodyclose.Analyzer,
		errcheck.Analyzer,
	}
}

// JoinAnalyzers помогает соединить слайсы с анализаторами в один.
func JoinAnalyzers(a ...[]*analysis.Analyzer) []*analysis.Analyzer {
	joined := make([]*analysis.Analyzer, 0, 100)
	for _, l := range a {
		joined = append(joined, l...)
	}
	return joined
}

// AllAnalyzers возвращает список всех подключаемых анализаторов.
func AllAnalyzers() []*analysis.Analyzer {
	return JoinAnalyzers(
		StaticChecks(),
		StandardPasses(),
		ThirdChecks(),
		Analyzers,
	)
}
