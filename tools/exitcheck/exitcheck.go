package exitcheck

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

const Doc = "Prohibits the use of `os.Exit` in the main func"

var Analyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  Doc,
	URL:  "https://github.com/nikimonax/go-metrics/tools/exitcheck",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" || strings.HasSuffix(pass.Pkg.Path(), ".test") {
		return nil, nil
	}

	var mainFunc *ast.FuncDecl

	for _, file := range pass.Files {
		if mainFunc = findFuncDecl(file, "main"); mainFunc != nil {
			break
		}
	}

	if mainFunc == nil {
		return nil, nil
	}

	osExitUsages := findCallExpr(pass, mainFunc)

	for _, osExitUsage := range osExitUsages {
		pass.ReportRangef(osExitUsage, `"os.Exit" usage`)
	}

	return nil, nil
}

func findFuncDecl(root ast.Node, funcName string) *ast.FuncDecl {
	var foundFuncDecl *ast.FuncDecl

	ast.Inspect(root, func(n ast.Node) bool {
		if foundFuncDecl != nil {
			return false
		}

		funcDecl, ok := n.(*ast.FuncDecl)

		if !ok {
			return true
		}

		if funcDecl.Name.Name == funcName {
			foundFuncDecl = funcDecl
			return false
		}

		return true
	})

	return foundFuncDecl
}

func findCallExpr(pass *analysis.Pass, root ast.Node) []*ast.CallExpr {
	foundCallExpr := make([]*ast.CallExpr, 0)

	ast.Inspect(root, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)

		if !ok {
			return true
		}

		if isExitCall(pass, callExpr) {
			foundCallExpr = append(foundCallExpr, callExpr)
		}

		return true
	})

	return foundCallExpr
}

func isExitCall(pass *analysis.Pass, callExpr *ast.CallExpr) bool {
	fn, ok := typeutil.Callee(pass.TypesInfo, callExpr).(*types.Func)

	if !ok {
		return false
	}

	pkg := fn.Pkg()

	return pkg != nil && pkg.Path() == "os" && fn.Name() == "Exit"
}
