// Package transpiler provides AssignToBitwisePass for converting bool assignments to bitwise operations
package transpiler

import (
	"go/ast"
	"go/token"
)

// AssignToBitwisePass converts bool field assignments to bitwise flag operations
// Transforms: cfg.Debug = true → cfg.flags |= FlagDebug
// Transforms: cfg.Debug = false → cfg.flags &^= FlagDebug
type AssignToBitwisePass struct{}

func (p *AssignToBitwisePass) Name() string {
	return "AssignToBitwise"
}

func (p *AssignToBitwisePass) Apply(file *ast.File, _ *token.FileSet, ctx *TranspileContext) error {
	ast.Inspect(file, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok || len(as.Lhs) != 1 || len(as.Rhs) != 1 {
			return true
		}

		sel, ok := as.Lhs[0].(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		valIdent, ok := as.Rhs[0].(*ast.Ident)
		if !ok {
			return true
		}

		for _, info := range ctx.Structs {
			if flagName, exists := info.FlagMapping[sel.Sel.Name]; exists {
				as.Lhs[0] = &ast.SelectorExpr{X: ident, Sel: ast.NewIdent("flags")}

				if valIdent.Name == "true" {
					as.Tok = token.OR_ASSIGN
					as.Rhs[0] = ast.NewIdent(flagName)
				} else if valIdent.Name == "false" {
					as.Tok = token.AND_NOT_ASSIGN
					as.Rhs[0] = ast.NewIdent(flagName)
				}
				break
			}
		}

		return true
	})
	return nil
}
