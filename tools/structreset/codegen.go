package structreset

import (
	"go/ast"
	"go/token"
	"go/types"
	"log/slog"
)

func generateFile(ctx codegenContext, structs []*ast.TypeSpec) *ast.File {
	decls := make([]ast.Decl, 0, len(structs))

	for _, typeSpec := range structs {
		structType := typeSpec.Type.(*ast.StructType)
		funcDecl := generateResetMethod(ctx.withTypeSpec(typeSpec), structType)
		decls = append(decls, funcDecl)
	}

	file := &ast.File{
		Name:  ast.NewIdent(ctx.pkg.Name),
		Decls: decls,
	}

	return file
}

func generateResetMethod(ctx codegenContext, strct *ast.StructType) *ast.FuncDecl {
	stmts := make([]ast.Stmt, 0, len(strct.Fields.List))

	for _, field := range strct.Fields.List {
		typ := ctx.typeOf(field.Type)

		for _, fieldName := range fieldNames(field) {
			sel := &ast.SelectorExpr{
				X:   ast.NewIdent(ReceiverIdent),
				Sel: ast.NewIdent(fieldName),
			}

			if stmt := generateResetStmt(ctx, sel, typ); stmt != nil {
				stmts = append(stmts, stmt)
			}
		}
	}

	return &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: ReceiverIdent}},
					Type:  &ast.StarExpr{X: ast.NewIdent(ctx.currentTypeSpec.Name.Name)},
				},
			},
		},
		Name: ast.NewIdent(MethodIdent),
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{List: stmts},
	}
}

func generateResetStmt(
	ctx codegenContext,
	x ast.Expr,
	typ types.Type,
) ast.Stmt {
	// 1. if any type resetable - reset it
	if ctx.isResetable(typ) {
		return generateResetCallStmt(ctx, x, typ)
	}

	switch under := typ.Underlying().(type) {
	// 2. basic types - zero assignment
	case *types.Basic:
		return generateZeroAssignStmt(ctx, x, under.Info())
	// 3. map - clear
	case *types.Map:
		return generateClearStmt(ctx, x)
	// 4. slice - cut
	case *types.Slice:
		return generateSliceCutAssignStmt(ctx, x)
	// 5. array - for loop reset
	case *types.Array:
		return generateForLoopResetStmt(ctx, x, under)
	// 6. pointer - nil check and reset
	case *types.Pointer:
		return generatePointerResetStmt(ctx, x, under)
	// 7. resetable structs, interfaces handled
	//    via Reset() in p.1 otherwise - skip
	case *types.Struct, *types.Interface:
	// 8. channels and functions - skip
	case *types.Chan, *types.Signature:

	default:
		slog.Warn("unhandled type", "type", typ.String())
	}

	return nil
}

// generates `if x != nil { x.Reset() }`
func generateResetCallStmt(
	_ codegenContext,
	x ast.Expr,
	typ types.Type,
) ast.Stmt {
	var stmt ast.Stmt = &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   x,
				Sel: ast.NewIdent(MethodIdent),
			},
		},
	}

	// guard
	if isNilable(typ) {
		stmt = nilCheckStmt(x, stmt)
	}

	return stmt
}

// generates `x = 0` / `x = ""` / `x = false`
func generateZeroAssignStmt(
	_ codegenContext,
	x ast.Expr,
	typeInfo types.BasicInfo,
) ast.Stmt {
	var rhs ast.Expr

	switch {
	case typeInfo&types.IsBoolean != 0:
		rhs = ast.NewIdent("false")
	case typeInfo&types.IsString != 0:
		rhs = &ast.BasicLit{Kind: token.STRING, Value: `""`}
	default:
		rhs = &ast.BasicLit{Kind: token.INT, Value: "0"}
	}

	return assignStmt(x, rhs)
}

// generates `x = x[:0]`
func generateSliceCutAssignStmt(
	_ codegenContext,
	x ast.Expr,
) ast.Stmt {
	rhs := &ast.SliceExpr{
		X:    x,
		High: &ast.BasicLit{Kind: token.INT, Value: "0"},
	}
	return assignStmt(x, rhs)
}

// generates `clear(x)`
func generateClearStmt(
	_ codegenContext,
	x ast.Expr,
) ast.Stmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun:  ast.NewIdent("clear"),
			Args: []ast.Expr{x},
		},
	}
}

// generates `for i := range x { /reset x[i]/ }`
func generateForLoopResetStmt(
	ctx codegenContext,
	x ast.Expr,
	typ *types.Array,
) ast.Stmt {
	loopVarIdent := ast.NewIdent("i")
	xWithIdx := &ast.IndexExpr{X: x, Index: loopVarIdent}
	resetStmt := generateResetStmt(ctx, xWithIdx, typ.Elem())

	if resetStmt == nil {
		return nil
	}

	return &ast.RangeStmt{
		Key:  loopVarIdent,
		Tok:  token.DEFINE,
		X:    x,
		Body: &ast.BlockStmt{List: []ast.Stmt{resetStmt}},
	}
}

// generates `if x != nil { /reset *x/ }`
func generatePointerResetStmt(
	ctx codegenContext,
	x ast.Expr,
	typ *types.Pointer,
) ast.Stmt {
	xWithStar := &ast.StarExpr{X: x}
	resetStmt := generateResetStmt(ctx, xWithStar, typ.Elem())

	if resetStmt == nil {
		return nil
	}

	return nilCheckStmt(x, resetStmt)
}

// helpers

func fieldNames(field *ast.Field) []string {
	// type Name struct {
	//     i, j T
	// }
	if len(field.Names) > 0 {
		names := make([]string, 0, len(field.Names))

		for _, name := range field.Names {
			names = append(names, name.Name)
		}

		return names
	}

	switch typ := field.Type.(type) {
	case *ast.Ident:
		// type Name struct {
		//     Embed
		// }
		return []string{typ.Name}
	case *ast.SelectorExpr:
		// type Name struct {
		//     pkg.Embed
		// }
		return []string{typ.Sel.Name}
	case *ast.StarExpr:
		// type Name struct {
		//     *Embed / *pkg.Embed
		// }
		return fieldNames(&ast.Field{Type: typ.X})
	default:
		return nil
	}
}

func isNilable(typ types.Type) bool {
	switch typ.Underlying().(type) {
	case *types.Pointer, *types.Interface:
		return true
	default:
		return false
	}
}

func assignStmt(lhs, rhs ast.Expr) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{lhs},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{rhs},
	}
}

func nilCheckStmt(x ast.Expr, stmt ast.Stmt) *ast.IfStmt {
	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  x,
			Op: token.NEQ,
			Y:  ast.NewIdent("nil"),
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{stmt}},
	}
}
