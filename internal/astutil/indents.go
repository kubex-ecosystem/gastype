package astutil

import "go/ast"

func GetRootIdent(expr ast.Expr) *ast.Ident {
	switch v := expr.(type) {
	case *ast.Ident:
		return v
	case *ast.SelectorExpr:
		return GetRootIdent(v.X)
	case *ast.StarExpr:
		return GetRootIdent(v.X)
	default:
		return nil
	}
}
