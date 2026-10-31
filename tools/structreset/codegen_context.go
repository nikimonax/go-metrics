package structreset

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type codegenContext struct {
	pkg        *packages.Package
	resetables map[types.Type]struct{}

	currentTypeSpec *ast.TypeSpec
}

func newCodegenContext(pkg *packages.Package, resetables map[types.Type]struct{}) codegenContext {
	return codegenContext{
		pkg:        pkg,
		resetables: resetables,
	}
}

func (ctx codegenContext) withTypeSpec(typeSpec *ast.TypeSpec) codegenContext {
	ctx.currentTypeSpec = typeSpec
	return ctx
}

func (ctx codegenContext) typeOf(expr ast.Expr) types.Type {
	if typ := ctx.pkg.TypesInfo.TypeOf(expr); typ != nil {
		return typ
	}

	ident, ok := expr.(*ast.Ident)

	if !ok {
		return types.Typ[types.Invalid]
	}

	obj, ok := ctx.pkg.TypesInfo.Uses[ident].(*types.TypeName)

	if !ok {
		return types.Typ[types.Invalid]
	}

	return obj.Type()
}

func (ctx codegenContext) isResetable(typ types.Type) bool {
	if hasResetMethod(typ) {
		return true
	}

	_, ok := ctx.resetables[typ]

	return ok
}

func hasResetMethod(typ types.Type) bool {
	return resetMethod(typ) != nil
}

func resetMethod(typ types.Type) *types.Func {
	fn := lookupReset(types.NewMethodSet(typ))

	if fn != nil {
		return fn
	}

	return lookupReset(types.NewMethodSet(types.NewPointer(typ)))
}

func lookupReset(methods *types.MethodSet) *types.Func {
	sel := methods.Lookup(nil, MethodIdent)

	if sel == nil {
		return nil
	}

	fn, ok := sel.Obj().(*types.Func)

	if !ok {
		return nil
	}

	sig := fn.Signature()

	if sig.Params().Len() != 0 || sig.Results().Len() != 0 {
		return nil
	}

	return fn
}
