package structreset

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"
)

func parseStruct(t *testing.T, src string) (*token.FileSet, *ast.StructType) {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(
		fset,
		"",
		"package tmp\n\n"+src,
		parser.SkipObjectResolution|parser.ParseComments,
	)

	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Decls, 1)
	require.IsType(t, file.Decls[0], new(ast.GenDecl))

	genDecl := file.Decls[0].(*ast.GenDecl)

	require.Len(t, genDecl.Specs, 1)
	require.IsType(t, genDecl.Specs[0], new(ast.TypeSpec))

	typeSpec := genDecl.Specs[0].(*ast.TypeSpec)

	require.IsType(t, typeSpec.Type, new(ast.StructType))

	return fset, typeSpec.Type.(*ast.StructType)
}

func TestFieldNames(t *testing.T) {
	src := `
	type Struct struct {
		Embed1
		pkg.Embed2
		*Embed3
		*pkg.Embed4
		x, y int
		str string
	} 
	`

	expectedNames := []string{
		"Embed1",
		"Embed2",
		"Embed3",
		"Embed4",
		"x",
		"y",
		"str",
	}
	actualNames := make([]string, 0, len(expectedNames))

	_, strct := parseStruct(t, src)
	for _, field := range strct.Fields.List {
		actualNames = append(actualNames, fieldNames(field)...)
	}

	require.ElementsMatch(t, expectedNames, actualNames)
}
