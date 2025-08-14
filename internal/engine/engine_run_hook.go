// Package transpiler provides a modular engine for Go AST transformations
package transpiler

import (
	"go/ast"
	"go/parser"
)

// Hook pequeno: use-o no teu Run() logo após parser.ParseFile para ativar o type-check.
func (e *Engine) afterParseHook(astFile *ast.File, filePath string) {
	// popula ctx.Info (types/selections) para que os passes usem info de tipo real
	_ = e.typeCheckFile(astFile)
}

// EXEMPLO de uso (no teu Engine.Run, após ParseFile):
// astFile, err := parser.ParseFile(e.Ctx.Fset, filePath, nil, parser.ParseComments)
// if err != nil { ... }
// e.afterParseHook(astFile, filePath)

