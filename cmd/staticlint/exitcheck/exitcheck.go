// Package exitcheck предоставляет статический анализатор, запрещающий
// прямой вызов os.Exit в функции main пакета main.
//
// Использование os.Exit в main препятствует корректному выполнению отложенных
// вызовов defer и затрудняет тестирование. Рекомендуется возвращать код выхода
// из main и вызывать os.Exit только в одной точке (например, в func main() { os.Exit(run()) }).
package exitcheck

import (
	"go/ast"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer — анализатор, проверяющий отсутствие прямых вызовов os.Exit в main пакета main.
var Analyzer = &analysis.Analyzer{
	Name:     "exitcheck",
	Doc:      "forbids direct calls to os.Exit in main function of main package",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// modulePath — префикс путей пакетов текущего проекта (чтобы не проверять зависимости).
const modulePath = "github.com/coolycow/shortener"

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}
	// Проверяем только пакеты текущего модуля, не зависимости
	if !strings.HasPrefix(pass.Pkg.Path(), modulePath) {
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	inspect.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}
		id, ok := sel.X.(*ast.Ident)
		if !ok {
			return
		}
		if id.Name != "os" || sel.Sel.Name != "Exit" {
			return
		}

		// Проверяем, что вызов находится внутри функции main (по файлу и вложенности)
		for _, f := range pass.Files {
			ast.Inspect(f, func(node ast.Node) bool {
				fn, ok := node.(*ast.FuncDecl)
				if !ok || fn.Name.Name != "main" || fn.Recv != nil || fn.Body == nil {
					return true
				}
				if pass.Fset.File(call.Pos()) != pass.Fset.File(fn.Pos()) {
					return true
				}
				// call внутри тела main? Файл должен быть из исходников проекта (не из go-build кэша).
				filePath := filepath.ToSlash(pass.Fset.File(call.Pos()).Name())
				if call.Pos() >= fn.Body.Pos() && call.Pos() <= fn.Body.End() &&
					strings.Contains(filePath, "shortener") && !strings.Contains(filePath, "go-build") {
					pass.Reportf(call.Pos(), "direct call to os.Exit in main is forbidden, use return and os.Exit in a single place")
					return false
				}
				return true
			})
		}
	})

	return nil, nil
}
