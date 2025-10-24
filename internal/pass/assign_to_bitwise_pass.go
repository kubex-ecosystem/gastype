// Package pass implements various AST transformation passes for Go code.
package pass

import (
	"go/ast"
	"go/token"

	"github.com/kubex-ecosystem/gastype/internal/astutil"
)

// AssignToBitwisePass converts bool field assignments to bitwise flag operations.
// Examples:
//
//	cfg.Debug = true  →  cfg.flags |= FlagDebug
//	cfg.Debug = false →  cfg.flags &^= FlagDebug
type AssignToBitwisePass struct{}

func NewAssignToBitwisePass() *AssignToBitwisePass {
	return &AssignToBitwisePass{}
}

func (p *AssignToBitwisePass) Name() string {
	return "AssignToBitwise"
}

func (p *AssignToBitwisePass) Apply(file *ast.File, fset *token.FileSet, ctx *astutil.TranspileContext) error {
	transformations := 0

	ast.Inspect(file, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok || len(as.Lhs) != 1 || len(as.Rhs) != 1 {
			return true
		}

		// Checa se LHS é algo como cfg.Debug
		sel, ok := as.Lhs[0].(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		// Checa se RHS é literal true/false
		valIdent, ok := as.Rhs[0].(*ast.Ident)
		if !ok {
			return true
		}

		// Procura se o campo é um mapeado para flag
		for _, info := range ctx.Structs {
			if flagName, exists := info.FlagMapping[sel.Sel.Name]; exists {
				// Substitui o campo pelo campo "flags"
				as.Lhs[0] = &ast.SelectorExpr{
					X:   ident,
					Sel: ast.NewIdent("flags"),
				}

				switch valIdent.Name {
				case "true":
					as.Tok = token.OR_ASSIGN
					as.Rhs[0] = ast.NewIdent(flagName)
				case "false":
					as.Tok = token.AND_NOT_ASSIGN
					as.Rhs[0] = ast.NewIdent(flagName)
				default:
					return true // não é booleano simples
				}

				transformations++

				// Registra no contexto para relatórios
				pos := fset.Position(as.Pos())
				ctx.RegisterAssignToBitwise(pos.Filename, sel.Sel.Name, flagName, valIdent.Name)

				break
			}
		}

		return true
	})

	if transformations > 0 {
		ctx.LogVerbose(nil, "🔄 AssignToBitwisePass: %d assignments converted", transformations)
	}

	return nil
}
