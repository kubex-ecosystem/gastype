package astutil

import (
	"go/ast"
	"go/types"
)

type Info struct {
	Types        map[ast.Expr]types.TypeAndValue
	Instances    map[*ast.Ident]types.Object
	Defs         map[*ast.Ident]types.Object
	Uses         map[*ast.Ident]types.Object
	Implicits    map[ast.Node]types.Object
	Selections   map[*ast.SelectorExpr]*types.Selection
	Scopes       map[ast.Node]*types.Scope
	InitOrder    []*types.Initializer
	FileVersions map[*ast.File]string
}

func NewInfo() *Info {
	return &Info{
		Types:        make(map[ast.Expr]types.TypeAndValue),
		Instances:    make(map[*ast.Ident]types.Object),
		Defs:         make(map[*ast.Ident]types.Object),
		Uses:         make(map[*ast.Ident]types.Object),
		Implicits:    make(map[ast.Node]types.Object),
		Selections:   make(map[*ast.SelectorExpr]*types.Selection),
		Scopes:       make(map[ast.Node]*types.Scope),
		InitOrder:    make([]*types.Initializer, 0),
		FileVersions: make(map[*ast.File]string),
	}
}

func (i *Info) GetTypes() map[ast.Expr]types.TypeAndValue { return i.Types }
func (i *Info) GetInstances() map[*ast.Ident]types.Object { return i.Instances }
func (i *Info) GetDefs() map[*ast.Ident]types.Object      { return i.Defs }
func (i *Info) GetUses() map[*ast.Ident]types.Object      { return i.Uses }
func (i *Info) GetImplicits() map[ast.Node]types.Object   { return i.Implicits }
func (i *Info) GetSelections() map[*ast.SelectorExpr]*types.Selection {
	return i.Selections
}
func (i *Info) GetScopes() map[ast.Node]*types.Scope  { return i.Scopes }
func (i *Info) GetInitOrder() []*types.Initializer    { return i.InitOrder }
func (i *Info) GetFileVersions() map[*ast.File]string { return i.FileVersions }

func (i *Info) SetTypes(types map[ast.Expr]types.TypeAndValue) { i.Types = types }
func (i *Info) SetInstances(instances map[*ast.Ident]types.Object) {
	i.Instances = instances
}
func (i *Info) SetDefs(defs map[*ast.Ident]types.Object) { i.Defs = defs }
func (i *Info) SetUses(uses map[*ast.Ident]types.Object) { i.Uses = uses }
func (i *Info) SetImplicits(implicits map[ast.Node]types.Object) {
	i.Implicits = implicits
}
func (i *Info) SetSelections(selections map[*ast.SelectorExpr]*types.Selection) {
	i.Selections = selections
}
func (i *Info) SetScopes(scopes map[ast.Node]*types.Scope)  { i.Scopes = scopes }
func (i *Info) SetInitOrder(initOrder []*types.Initializer) { i.InitOrder = initOrder }
func (i *Info) SetFileVersions(fileVersions map[*ast.File]string) {
	i.FileVersions = fileVersions
}

func (i *Info) AddType(key ast.Expr, value types.TypeAndValue) {
	i.Types[key] = value
}
func (i *Info) AddInstance(key *ast.Ident, value types.Object) {
	i.Instances[key] = value
}
func (i *Info) AddDef(key *ast.Ident, value types.Object) {
	i.Defs[key] = value
}
func (i *Info) AddUse(key *ast.Ident, value types.Object) {
	i.Uses[key] = value
}
func (i *Info) AddImplicit(key ast.Node, value types.Object) {
	i.Implicits[key] = value
}
func (i *Info) AddSelection(key *ast.SelectorExpr, value *types.Selection) {
	i.Selections[key] = value
}
func (i *Info) AddScope(key ast.Node, value *types.Scope) { i.Scopes[key] = value }
func (i *Info) AddInitOrder(value *types.Initializer)     { i.InitOrder = append(i.InitOrder, value) }

func (i *Info) AddFileVersion(key *ast.File, value string) {
	i.FileVersions[key] = value
}
func (i *Info) GetFileVersion(key *ast.File) string {
	return i.FileVersions[key]
}
func (i *Info) RemoveFileVersion(key *ast.File) { delete(i.FileVersions, key) }
func (i *Info) ClearFileVersions()              { i.FileVersions = make(map[*ast.File]string) }
func (i *Info) Clear() {
	i.Types = make(map[ast.Expr]types.TypeAndValue)
	i.Instances = make(map[*ast.Ident]types.Object)
	i.Defs = make(map[*ast.Ident]types.Object)
	i.Uses = make(map[*ast.Ident]types.Object)
	i.Implicits = make(map[ast.Node]types.Object)
	i.Selections = make(map[*ast.SelectorExpr]*types.Selection)
	i.Scopes = make(map[ast.Node]*types.Scope)
	i.InitOrder = make([]*types.Initializer, 0)
	i.FileVersions = make(map[*ast.File]string)
}
