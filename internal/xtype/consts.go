// Package xtype contém utilitários para trabalhar com tipos no Go.
package xtype

import (
	"go/ast"
	"go/constant"
	"go/types"
)

type ResolvedConst struct {
	Type  types.Type
	Value constant.Value // nil se não for const
	From  string         // "Types"|"Uses"|"Defs"|"Uses(sel)"
}

func ResolveConst(info *types.Info, expr ast.Expr) (ResolvedConst, bool) {
	// 1) info.Types (literal/expr folded)
	if tv, ok := info.Types[expr]; ok && tv.Type != nil {
		return ResolvedConst{Type: tv.Type, Value: tv.Value, From: "Types"}, true
	}
	// 2) Ident -> Uses/Defs
	if id, ok := expr.(*ast.Ident); ok && id != nil {
		if obj := info.Uses[id]; obj != nil {
			if c, ok := obj.(*types.Const); ok {
				return ResolvedConst{Type: c.Type(), Value: c.Val(), From: "Uses"}, true
			}
			return ResolvedConst{Type: obj.Type(), Value: nil, From: "Uses"}, true
		}
		if obj := info.Defs[id]; obj != nil {
			if c, ok := obj.(*types.Const); ok {
				return ResolvedConst{Type: c.Type(), Value: c.Val(), From: "Defs"}, true
			}
			return ResolvedConst{Type: obj.Type(), Value: nil, From: "Defs"}, true
		}
	}
	// 3) SelectorExpr pkg.X
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		id := sel.Sel
		if id != nil {
			if obj := info.Uses[id]; obj != nil {
				if c, ok := obj.(*types.Const); ok {
					return ResolvedConst{Type: c.Type(), Value: c.Val(), From: "Uses(sel)"}, true
				}
				return ResolvedConst{Type: obj.Type(), Value: nil, From: "Uses(sel)"}, true
			}
		}
	}
	return ResolvedConst{}, false
}

func IsConstString(info *types.Info, expr ast.Expr) bool {
	rc, ok := ResolveConst(info, expr)
	if !ok {
		return false
	}
	if rc.Value != nil && rc.Value.Kind() == constant.String {
		return true
	}
	t := rc.Type
	if t == nil {
		return false
	}
	// default untyped
	if b, ok := t.Underlying().(*types.Basic); ok && (b.Info()&types.IsUntyped) != 0 {
		t = types.Default(t)
	}
	return types.Identical(t, types.Typ[types.String])
}
