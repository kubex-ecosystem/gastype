// Package transpiler provides a modular engine for Go AST transformations
package transpiler

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/types"

	"github.com/rafa-mori/gastype/internal/astutil"

	gl "github.com/rafa-mori/gastype/internal/module/logger"
)

// typeCheckFile popula ctx.Info (types/selections/scopes) para o arquivo atual.
// Chame isso logo após parser.ParseFile(...) dentro do loop do Engine.Run.
func (e *Engine) typeCheckFile(astFile *ast.File) error {
	if e.Ctx.Info == nil {
		e.Ctx.Info = astutil.NewInfo()
	}
	info := &types.Info{
		Types:      e.Ctx.GetTypes(),
		Defs:       e.Ctx.GetDefs(),
		Uses:       e.Ctx.GetUses(),
		Selections: e.Ctx.GetSelections(),
		Scopes:     e.Ctx.GetScopes(),
	}
	conf := types.Config{Importer: importer.Default()}
	_, err := conf.Check("", e.Ctx.Fset, []*ast.File{astFile}, info)
	if err != nil {
		// não quebra pipeline; só loga o warning
		gl.Log("warn", fmt.Sprintf("type-check warnings: %v", err))
	}
	return nil
}
