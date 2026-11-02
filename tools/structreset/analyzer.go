package structreset

import (
	"go/ast"
	"go/token"
	"maps"
	"reflect"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const Doc = "Finds and validates `generate:reset` usages"

const Directive = "generate:reset"

var Analyzer = &analysis.Analyzer{
	Name:       "structreset",
	Doc:        Doc,
	URL:        "https://github.com/nikimonax/go-metrics/tools/structreset",
	Run:        run,
	ResultType: reflect.TypeFor[[]*ast.TypeSpec](),
}

func run(pass *analysis.Pass) (any, error) {
	result := make([]*ast.TypeSpec, 0)

	for _, file := range pass.Files {
		typeSpecs := processFile(pass, file)
		result = append(result, typeSpecs...)
	}

	return result, nil
}

func processFile(pass *analysis.Pass, file *ast.File) []*ast.TypeSpec {
	typeSpecs := make(map[*ast.CommentGroup]*ast.TypeSpec)

	// collect top-level structs with generate:reset comment
	for _, decl := range file.Decls {
		// GenSpec -> []TypeSpec -> StructType
		genDecl, ok := decl.(*ast.GenDecl)

		if !ok {
			continue
		}

		if genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)

			if !ok {
				continue
			}

			_, ok = typeSpec.Type.(*ast.StructType)

			if !ok {
				continue
			}

			doc := typeSpec.Doc

			if doc == nil {
				doc = genDecl.Doc
			}

			if doc == nil {
				continue
			}

			if !hasDirectiveComment(doc) {
				continue
			}

			typeSpecs[doc] = typeSpec
		}
	}

	// find all incorrect `generate:reset` usages
	for _, group := range file.Comments {
		comment := directiveComment(group)

		if comment == nil {
			continue
		}

		if _, ok := typeSpecs[group]; ok {
			continue
		}

		pass.ReportRangef(comment, `Using "%s" is not applicable here`, Directive)
	}

	return slices.Collect(maps.Values(typeSpecs))
}

func hasDirectiveComment(group *ast.CommentGroup) bool {
	return directiveComment(group) != nil
}

func directiveComment(group *ast.CommentGroup) *ast.Comment {
	for _, comment := range group.List {
		text := comment.Text

		if strings.HasPrefix(text, "//") {
			text = strings.TrimPrefix(text, "//")
		} else if strings.HasPrefix(text, "/*") {
			text = strings.TrimPrefix(text, "/*")
			text = strings.TrimSuffix(text, "*/")
		}

		fields := strings.Fields(text)

		if len(fields) > 0 && fields[0] == Directive {
			return comment
		}
	}

	return nil
}
