// Package pass implements various AST transformation passes for the Gastype project.
package pass

import (
	"go/ast"
	"go/token"

	"github.com/rafa-mori/gastype/internal/astutil"
)

// FieldAccessToBitwisePass transforms any field access to bitwise operations
// This covers:
// - Function arguments
// - Return values
// - Assignments
// - Standalone expressions
type FieldAccessToBitwisePass struct{}

func NewFieldAccessToBitwisePass() *FieldAccessToBitwisePass {
	return &FieldAccessToBitwisePass{}
}

func (p *FieldAccessToBitwisePass) Name() string {
	return "FieldAccessToBitwise"
}

func (p *FieldAccessToBitwisePass) Apply(file *ast.File, fset *token.FileSet, ctx *astutil.TranspileContext) error {
	transformations := 0

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr: // function arguments
			for i, arg := range node.Args {
				if newArg := p.transformSelectorExpr(arg, ctx); newArg != nil {
					node.Args[i] = newArg
					transformations++
				}
			}

		case *ast.ReturnStmt: // return values
			for i, result := range node.Results {
				if newResult := p.transformSelectorExpr(result, ctx); newResult != nil {
					node.Results[i] = newResult
					transformations++
				}
			}

		case *ast.AssignStmt: // assignments
			for i, rhs := range node.Rhs {
				if newRhs := p.transformSelectorExpr(rhs, ctx); newRhs != nil {
					node.Rhs[i] = newRhs
					transformations++
				}
			}

		case *ast.ExprStmt: // standalone expressions (rare but possible)
			if newX := p.transformSelectorExpr(node.X, ctx); newX != nil {
				node.X = newX
				transformations++
			}
		}
		return true
	})

	if transformations > 0 {
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

	// Ensure we are accessing a struct field that was mapped to flags
	if info, exists := ctx.Structs[sel.X.(*ast.Ident).Name]; exists {
		if flagName, ok := info.FlagMapping[sel.Sel.Name]; ok {
			// ✅ Transform to: (obj.flags & FlagStruct_Field) != 0
			return &ast.BinaryExpr{
				X: &ast.ParenExpr{
					X: &ast.BinaryExpr{
						X: &ast.SelectorExpr{
							X:   sel.X, // preserve original object
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
