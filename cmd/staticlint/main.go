// Package main запускает multichecker — статический анализатор кода на Go,
// объединяющий стандартные анализаторы go/analysis/passes, анализаторы staticcheck
// (класс SA и другие), сторонние анализаторы и собственный exitcheck.
//
// Запуск:
//
//	go run ./cmd/staticlint/ ./...
//	go build -o staticlint ./cmd/staticlint && ./staticlint ./...
//
// Флаги: поддерживаются стандартные флаги go/analysis (например -json, -fix).
// Для включения/выключения конкретного анализатора используйте -NAME (включить) или -NAME=false.
package main

import (
	"github.com/coolycow/gophkeeper/cmd/staticlint/exitcheck"
	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	analyzers := []*analysis.Analyzer{
		// Стандартные анализаторы golang.org/x/tools/go/analysis/passes
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		deepequalerrors.Analyzer,
		defers.Analyzer,
		directive.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		// Собственный анализатор: запрет os.Exit в main
		exitcheck.Analyzer,
		// Сторонние анализаторы
		errcheck.Analyzer,
		bodyclose.Analyzer,
	}

	// Все SA-анализаторы staticcheck
	for _, a := range staticcheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	// Анализаторы других классов staticcheck: simple (S), stylecheck (ST)
	for _, a := range simple.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}
	for _, a := range stylecheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	multichecker.Main(analyzers...)
}
