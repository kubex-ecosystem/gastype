package pass

import (
	"go/ast"
	"go/token"

	"github.com/rafa-mori/gastype/internal/astutil"
)

// FieldAccessToBitwisePass transforms any field access to bitwise operations
// REVOLUTIONARY: Catches field access in function arguments, returns, etc!
type FieldAccessToBitwisePass struct{}

func NewFieldAccessToBitwisePass() *FieldAccessToBitwisePass {
	return &FieldAccessToBitwisePass{}
}

func (p *FieldAccessToBitwisePass) Name() string {
	return "FieldAccessToBitwise"
}

func (p *FieldAccessToBitwisePass) Apply(file *ast.File, fset *token.FileSet, ctx *astutil.TranspileContext) error {
	transformations := 0

	// We need to transform expressions in place by replacing in parent nodes
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			// Transform function arguments
			for i, arg := range node.Args {
				if newArg := p.transformSelectorExpr(arg, ctx); newArg != nil {
					node.Args[i] = newArg
					transformations++
				}
			}
		case *ast.ReturnStmt:
			// Transform return values
			for i, result := range node.Results {
				if newResult := p.transformSelectorExpr(result, ctx); newResult != nil {
					node.Results[i] = newResult
					transformations++
				}
			}
		case *ast.AssignStmt:
			// Transform right-hand side of assignments
			for i, rhs := range node.Rhs {
				if newRhs := p.transformSelectorExpr(rhs, ctx); newRhs != nil {
					node.Rhs[i] = newRhs
					transformations++
				}
			}
		}
		return true
	})

	if transformations > 0 {
		// Simple log for now - could be enhanced with proper logging
		ctx.LogVerbose(fset, "🔄 FieldAccessToBitwisePass: %d transformations applied", transformations)
	}

	return nil
}

// transformSelectorExpr checks if an expression is a selector that needs transformation
func (p *FieldAccessToBitwisePass) transformSelectorExpr(expr ast.Expr, ctx *astutil.TranspileContext) ast.Expr {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return nil
	}

	// Check if this is a bool field access that was mapped to flags
	for _, info := range ctx.Structs {
		if flagName, exists := info.FlagMapping[sel.Sel.Name]; exists {
			// 🚀 REVOLUTIONARY: Transform cfg.Debug → (cfg.flags & FlagConfig_Debug) != 0
			return &ast.BinaryExpr{
				X: &ast.ParenExpr{
					X: &ast.BinaryExpr{
						X: &ast.SelectorExpr{
							X:   ident,
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
		}
	}

	return nil
}
