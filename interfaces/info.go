package interfaces

import (
	"go/ast"
	"go/types"
)

type IInfo interface {
	GetTypes() map[ast.Expr]types.TypeAndValue
	GetInstances() map[*ast.Ident]types.Object
	GetDefs() map[*ast.Ident]types.Object
	GetUses() map[*ast.Ident]types.Object
	GetImplicits() map[ast.Node]types.Object
	GetSelections() map[*ast.SelectorExpr]*types.Selection
	GetScopes() map[ast.Node]*types.Scope
	GetInitOrder() []*types.Initializer
	GetFileVersions() map[*ast.File]string

	SetTypes(types map[ast.Expr]types.TypeAndValue)
	SetInstances(instances map[*ast.Ident]types.Object)
	SetDefs(defs map[*ast.Ident]types.Object)
	SetUses(uses map[*ast.Ident]types.Object)
	SetImplicits(implicits map[ast.Node]types.Object)
	SetSelections(selections map[*ast.SelectorExpr]*types.Selection)
	SetScopes(scopes map[ast.Node]*types.Scope)
	SetInitOrder(initOrder []*types.Initializer)
	SetFileVersions(fileVersions map[*ast.File]string)

	AddType(key ast.Expr, value types.TypeAndValue)
	AddInstance(key *ast.Ident, value types.Object)
	AddDef(key *ast.Ident, value types.Object)
	AddUse(key *ast.Ident, value types.Object)
	AddImplicit(key ast.Node, value types.Object)
	AddSelection(key *ast.SelectorExpr, value *types.Selection)
	AddScope(key ast.Node, value *types.Scope)
	AddInitOrder(value *types.Initializer)
	AddFileVersion(key *ast.File, value string)

	GetFileVersion(key *ast.File) string

	RemoveFileVersion(key *ast.File)
	ClearFileVersions()
	Clear()
}
