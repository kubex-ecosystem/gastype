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

func IsValidIdent(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if !IsValidIdentRune(r) {
			return false
		}
	}
	return true
}

func IsValidIdentRune(r rune) bool {
	return r == '_' || r == '$' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}
