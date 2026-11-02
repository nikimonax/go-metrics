package structreset

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"reflect"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/checker"
	"golang.org/x/tools/go/packages"
)

func Main(patterns ...string) error {
	cfg := packages.Config{Mode: packages.LoadAllSyntax}

	pkgs, err := packages.Load(&cfg, patterns...)

	if err != nil {
		return err
	}

	generated, err := Generate(pkgs)

	if err != nil {
		return err
	}

	for pkg, file := range generated {
		filePath := filepath.Join(pkg.Dir, GeneratedFileName)

		if err := writeFile(file, filePath); err != nil {
			return err
		}
	}

	return nil
}

func Generate(pkgs []*packages.Package) (map[*packages.Package]*ast.File, error) {
	graph, err := checker.Analyze([]*analysis.Analyzer{Analyzer}, pkgs, nil)

	if err != nil {
		return nil, err
	}

	type task struct {
		pkg     *packages.Package
		structs []*ast.TypeSpec
	}

	result := make(map[*packages.Package]*ast.File)
	resettables := make(map[types.Type]struct{})

	tasks := make([]task, 0)

	for act := range graph.All() {
		if act.Result == nil {
			continue
		}

		structs, ok := act.Result.([]*ast.TypeSpec)

		if !ok {
			return nil, fmt.Errorf(
				"the analyzer returned an unexpected type '%s', but expected '%s'",
				reflect.TypeOf(structs),
				reflect.TypeFor[[]*ast.TypeSpec](),
			)
		}

		if len(structs) == 0 {
			continue
		}

		for _, typeSpec := range structs {
			resettables[act.Package.TypesInfo.TypeOf(typeSpec.Name)] = struct{}{}
		}

		tasks = append(tasks, task{pkg: act.Package, structs: structs})
	}

	for _, task := range tasks {
		ctx := newCodegenContext(task.pkg, resettables)
		result[task.pkg] = generateFile(ctx, task.structs)
	}

	return result, nil
}

func writeFile(fileNode *ast.File, path string) (err error) {
	file, err := os.Create(path)

	if err != nil {
		return err
	}

	defer func() { err = errors.Join(err, file.Close()) }()

	if _, err := file.WriteString(GeneratedFileTitle); err != nil {
		return err
	}

	fset := token.NewFileSet()
	return format.Node(file, fset, fileNode)
}
