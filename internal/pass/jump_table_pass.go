// Package pass provides JumpTablePass for optimizing if-chains to jump tables
package pass

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/rafa-mori/gastype/internal/astutil"
)

// JumpTablePass converts chained if-else statements to jump table optimizations
// Transforms: if cmd == "start" { startService() } else if cmd == "stop" { stopService() }
// To: var cmdTable = map[string]func(){"start": startService, "stop": stopService}; if fn, ok := cmdTable[cmd]; ok { fn() }
type JumpTablePass struct{}

func NewJumpTablePass() *JumpTablePass {
	return &JumpTablePass{}
}

func (p *JumpTablePass) Name() string {
	return "JumpTable"
}

func (p *JumpTablePass) Apply(file *ast.File, fset *token.FileSet, ctx *astutil.TranspileContext) error {
	transformations := 0
	candidates := 0

	ast.Inspect(file, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}

		// Check if this is a binary expression with == comparison
		binExpr, ok := ifStmt.Cond.(*ast.BinaryExpr)
		if !ok || binExpr.Op != token.EQL {
			return true
		}

		// For now, we'll just identify candidates for jump table optimization
		// Full implementation would require more complex AST manipulation

		// Check if left side is an identifier and right side is a string literal
		leftIdent, leftOk := binExpr.X.(*ast.Ident)
		rightLit, rightOk := binExpr.Y.(*ast.BasicLit)

		if leftOk && rightOk && rightLit.Kind == token.STRING {
			candidates++
			fmt.Printf("    🚀 Found jump table candidate: %s == %s\n", leftIdent.Name, rightLit.Value)

			// Count chained else-if statements
			chainLength := 1
			current := ifStmt
			for current.Else != nil {
				if elseIf, ok := current.Else.(*ast.IfStmt); ok {
					if binExpr, ok := elseIf.Cond.(*ast.BinaryExpr); ok && binExpr.Op == token.EQL {
						if leftIdent2, ok := binExpr.X.(*ast.Ident); ok && leftIdent2.Name == leftIdent.Name {
							chainLength++
							current = elseIf
							continue
						}
					}
				}
				break
			}

			if chainLength >= 3 { // Only optimize chains with 3+ conditions
				fmt.Printf("    ⚡ Chain length %d - good candidate for jump table optimization\n", chainLength)
				transformations++

				// TODO: Implement actual jump table generation
				// This would involve:
				// 1. Creating a map[string]func() variable
				// 2. Extracting function calls from each branch
				// 3. Replacing the if-chain with map lookup

				// For now, just mark it as optimizable
			}
		}

		return true
	})

	if transformations > 0 {
		fmt.Printf("  🔄 JumpTablePass: %d optimizable chains found (%d total candidates)\n", transformations, candidates)
	}

	return nil
}
