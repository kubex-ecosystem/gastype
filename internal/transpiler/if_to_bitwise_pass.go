// Package transpiler provides IfToBitwisePass for converting bool conditions to bitwise checks
package transpiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// IfToBitwisePass converts bool field conditions to bitwise flag checks
// Transforms: if cfg.Debug → if cfg.flags & FlagDebug != 0
type IfToBitwisePass struct{}

func (p *IfToBitwisePass) Name() string {
	return "IfToBitwise"
}

func (p *IfToBitwisePass) Apply(file *ast.File, fset *token.FileSet, ctx *TranspileContext) error {
	transformations := 0

	ast.Inspect(file, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}

		// Check if condition is a selector expression (e.g., cfg.Debug)
		sel, ok := ifStmt.Cond.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		// Look for matching bool fields that were converted to flags
		for structName, info := range ctx.Structs {
			for _, field := range info.BoolFields {
				if sel.Sel.Name == field {
					// Create package-scoped flag name to avoid collisions across packages
					pkgName := file.Name.Name
					flagName := fmt.Sprintf("Flag%s_%s_%s", pkgName, structName, strings.Title(field))

					// Replace condition with bitwise check: cfg.flags & FlagPkg_Config_Debug != 0
					bitwiseCheck := &ast.BinaryExpr{
						X: &ast.BinaryExpr{
							X:  &ast.SelectorExpr{X: ident, Sel: ast.NewIdent("flags")},
							Op: token.AND,
							Y:  ast.NewIdent(flagName),
						},
						Op: token.NEQ,
						Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
					}

					ifStmt.Cond = bitwiseCheck
					transformations++

					fmt.Printf("    ⚡ Transformed if %s.%s → bitwise check\n", ident.Name, field)

					// Track transformation
					if info.Transformations == nil {
						info.Transformations = make(map[string]string)
					}
					info.Transformations["if_"+field] = "bitwise_check"

					// Update struct info in context
					ctx.Structs[structName] = info
				}
			}
		}
		return true
	})

	if transformations > 0 {
		fmt.Printf("  🔄 IfToBitwisePass: %d transformations applied\n", transformations)
	}

	return nil
}
