package astutil

import (
	"go/ast"
	"go/token"
)

// NewBitwiseCheck Cria uma expressão bitwise: flags & FlagXYZ != 0
func NewBitwiseCheck(objName, flagsField, flagConst string) ast.Expr {
	return &ast.BinaryExpr{
		X: &ast.BinaryExpr{
			X:  &ast.SelectorExpr{X: ast.NewIdent(objName), Sel: ast.NewIdent(flagsField)},
			Op: token.AND,
			Y:  ast.NewIdent(flagConst),
		},
		Op: token.NEQ,
		Y:  ast.NewIdent("0"),
	}
}

// NewConst Cria uma constante AST
func NewConst(name, typ, value string) *ast.GenDecl {
	return &ast.GenDecl{
		Tok: token.CONST,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names:  []*ast.Ident{ast.NewIdent(name)},
				Type:   ast.NewIdent(typ),
				Values: []ast.Expr{ast.NewIdent(value)},
			},
		},
	}
}
