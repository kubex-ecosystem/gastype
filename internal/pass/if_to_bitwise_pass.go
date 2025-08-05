// Package pass implements various code transformations for Go ASTs.
package pass

import (
	"go/ast"
	"go/token"

	"github.com/rafa-mori/gastype/internal/astutil"
)

// IfToBitwisePass converts bool field conditions in if-statements to bitwise flag checks
// Example:
//
//	if cfg.Debug { ... }
//
// becomes:
//
//	if (cfg.flags & FlagConfig_Debug) != 0 { ... }
type IfToBitwisePass struct{}

func NewIfToBitwisePass() *IfToBitwisePass {
	return &IfToBitwisePass{}
}

func (p *IfToBitwisePass) Name() string {
	return "IfToBitwise"
}

func (p *IfToBitwisePass) Apply(file *ast.File, _ *token.FileSet, ctx *astutil.TranspileContext) error {
	transformations := 0

	ast.Inspect(file, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}

		// Verifica se a condição é um acesso a campo (SelectorExpr)
		sel, ok := ifStmt.Cond.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		// O X do seletor pode ser Ident (cfg) ou algo mais complexo (ex: getCfg())
		structIdent, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		// Busca no contexto se a struct tem esse campo mapeado para flag
		if info, exists := ctx.Structs[structIdent.Name]; exists {
			if flagName, ok := info.FlagMapping[sel.Sel.Name]; ok {
				// Substitui a condição do if para a checagem bitwise
				ifStmt.Cond = &ast.BinaryExpr{
					X: &ast.ParenExpr{
						X: &ast.BinaryExpr{
							X: &ast.SelectorExpr{
								X:   sel.X, // preserva o objeto original
								Sel: ast.NewIdent("flags"),
							},
							Op: token.AND,
							Y:  ast.NewIdent(flagName),
						},
					},
					Op: token.NEQ,
					Y: &ast.BasicLit{
						Kind:  token.INT,
						Value: "0",
					},
				}
				transformations++
			}
		}

		return true
	})

	if transformations > 0 {
		ctx.LogVerbose(nil, "⚡ IfToBitwisePass: %d transformations applied", transformations)
	}

	return nil
}
