// Package pass implements various code transformations for Go ASTs.
package pass

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/kubex-ecosystem/gastype/internal/astutil"
)

// IfToBitwisePass cfg.Debug  ->  (cfg.flags & FlagConfig_Debug) != 0
type IfToBitwisePass struct{}

func NewIfToBitwisePass() *IfToBitwisePass { return &IfToBitwisePass{} }
func (p *IfToBitwisePass) Name() string    { return "IfToBitwise" }

func (p *IfToBitwisePass) Apply(file *ast.File, _ *token.FileSet, ctx *astutil.TranspileContext) error {
	transformations := 0

	ast.Inspect(file, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		sel, ok := ifStmt.Cond.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		// Usa types.Selections: pegamos o campo e o tipo receptor reais
		selInfo := ctx.GetSelections()[sel]
		if selInfo == nil || selInfo.Obj() == nil {
			return true
		}

		fieldName := selInfo.Obj().Name()
		recv := selInfo.Recv()
		if recv == nil {
			return true
		}

		typName := strings.TrimPrefix(recv.String(), "*")

		// Agora lookup por TIPO (não por nome de variável)
		if info, ok := ctx.Structs[typName]; ok {
			if flagName, ok := info.FlagMapping[fieldName]; ok {
				ifStmt.Cond = &ast.BinaryExpr{
					X: &ast.ParenExpr{
						X: &ast.BinaryExpr{
							X:  &ast.SelectorExpr{X: sel.X, Sel: ast.NewIdent("flags")},
							Op: token.AND,
							Y:  ast.NewIdent(flagName),
						},
					},
					Op: token.NEQ,
					Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
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
