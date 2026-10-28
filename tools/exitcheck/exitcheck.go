package exitcheck

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
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

	osExitUsages := findCallExpr(mainFunc, "os.Exit")

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

func findCallExpr(root ast.Node, funcName string) []*ast.CallExpr {
	foundCallExpr := make([]*ast.CallExpr, 0)

	ast.Inspect(root, func(n ast.Node) bool {
		callDecl, ok := n.(*ast.CallExpr)

		if !ok {
			return true
		}

		if getFuncName(callDecl.Fun) == funcName {
			foundCallExpr = append(foundCallExpr, callDecl)
		}

		return true
	})

	return foundCallExpr
}

func getFuncName(expr ast.Expr) string {
	switch typedExpr := expr.(type) {
	case *ast.Ident:
		// Regular function call (e.g., calculate(5, 10))
		return typedExpr.Name

	case *ast.SelectorExpr:
		// Packaged function or method call (e.g., fmt.Println("..."))
		if pkgIdent, ok := typedExpr.X.(*ast.Ident); ok {
			return pkgIdent.Name + "." + typedExpr.Sel.Name
		}
		return typedExpr.Sel.Name

	default:
		return ""
	}
}
