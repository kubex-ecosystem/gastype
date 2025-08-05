// Package transpiler provides AssignToBitwisePass for converting bool assignments to bitwise operations
package transpiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// AssignToBitwisePass converts bool field assignments to bitwise flag operations
// Transforms: cfg.Debug = true → cfg.flags |= FlagDebug
// Transforms: cfg.Debug = false → cfg.flags &^= FlagDebug
type AssignToBitwisePass struct{}

func (p *AssignToBitwisePass) Name() string {
	return "AssignToBitwise"
}

func (p *AssignToBitwisePass) Apply(file *ast.File, fset *token.FileSet, ctx *TranspileContext) error {
	transformations := 0

	ast.Inspect(file, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok || len(as.Lhs) != 1 || len(as.Rhs) != 1 {
			return true
		}

		// Check if left side is a selector expression (e.g., cfg.Debug)
		sel, ok := as.Lhs[0].(*ast.SelectorExpr)
		if !ok {
			return true
		}

		// Check if right side is a boolean literal
		valIdent, ok := as.Rhs[0].(*ast.Ident)
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

					if valIdent.Name == "true" {
						// cfg.flags |= FlagX
						as.Lhs[0] = &ast.SelectorExpr{X: ident, Sel: ast.NewIdent("flags")}
						as.Tok = token.OR_ASSIGN
						as.Rhs[0] = ast.NewIdent(flagName)

						fmt.Printf("    ⚡ Transformed %s.%s = true → flags |= %s\n", ident.Name, field, flagName)

					} else if valIdent.Name == "false" {
						// cfg.flags &^= FlagX (AND NOT assignment)
						as.Lhs[0] = &ast.SelectorExpr{X: ident, Sel: ast.NewIdent("flags")}
						as.Tok = token.AND_NOT_ASSIGN
						as.Rhs[0] = ast.NewIdent(flagName)

						fmt.Printf("    ⚡ Transformed %s.%s = false → flags &^= %s\n", ident.Name, field, flagName)
					}

					transformations++

					// Track transformation
					if info.Transformations == nil {
						info.Transformations = make(map[string]string)
					}
					info.Transformations["assign_"+field] = "bitwise_operation"

					// Update struct info in context
					ctx.Structs[structName] = info
				}
			}
		}
		return true
	})

	if transformations > 0 {
		fmt.Printf("  🔄 AssignToBitwisePass: %d transformations applied\n", transformations)
	}

	return nil
}
