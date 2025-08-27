package astutil

import "go/ast"

// ReplaceNode replaces a target AST node with a new one.
// Returns true if the replacement was applied.
func ReplaceNode(file *ast.File, target ast.Node, replacement ast.Node) bool {
	replaced := false
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.BlockStmt:
			for i, stmt := range node.List {
				if stmt == target {
					node.List[i] = replacement.(ast.Stmt)
					replaced = true
					return false
				}
			}
		}
		return true
	})
	return replaced
}
