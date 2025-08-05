// Package transpiler provides IfToBitwisePass for converting bool conditions to bitwise checks
package transpiler

import (
	"go/ast"
	"go/token"

	"github.com/rafa-mori/gastype/internal/astutil"
)

// IfToBitwisePass converts bool field conditions to bitwise flag checks
// Transforms: if cfg.Debug → if cfg.flags & FlagDebug != 0
type IfToBitwisePass struct{}

func (p *IfToBitwisePass) Name() string {
	return "IfToBitwise"
}

func (p *IfToBitwisePass) Apply(file *ast.File, _ *token.FileSet, ctx *astutil.TranspileContext) error {
	ast.Inspect(file, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}

		sel, ok := ifStmt.Cond.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		// Descobre struct e flag name
		for _, info := range ctx.Structs {
			if flagName, exists := info.FlagMapping[sel.Sel.Name]; exists {
				ifStmt.Cond = &ast.BinaryExpr{
					X: &ast.BinaryExpr{
						X:  &ast.SelectorExpr{X: ident, Sel: ast.NewIdent("flags")},
						Op: token.AND,
						Y:  ast.NewIdent(flagName),
					},
					Op: token.NEQ,
					Y:  ast.NewIdent("0"),
				}
				break
			}
		}

		return true
	})
	return nil
}
