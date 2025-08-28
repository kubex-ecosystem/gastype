// Package xast contém utilitários para trabalhar com a AST (Abstract Syntax Tree) do Go.
package xast

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
)

func InImportOrTag(c *Cursor) bool {
	switch n := c.Node().(type) {
	case *ast.BasicLit:
		// Import path
		if imp, ok := c.Parent().(*ast.ImportSpec); ok && imp.Path == n {
			return true
		}
		// Struct tag
		if fld, ok := c.Parent().(*ast.Field); ok && fld.Tag == n {
			return true
		}
	}
	return false
}

// ResolveLocalStringType tenta preservar um tipo nominal local (Ident/Selector) se presente.
// 1) pai é CallExpr de conversão -> usa Fun
// 2) pai/avô é ValueSpec com Type -> usa Type
// 3) fallback: "string"
func ResolveLocalStringType(c *Cursor, _ *types.Info) ast.Expr {
	// (1) conversão explícita
	if call, ok := c.Parent().(*ast.CallExpr); ok && call.Fun != nil {
		switch call.Fun.(type) {
		case *ast.Ident, *ast.SelectorExpr:
			return call.Fun
		}
	}
	// (2) declaração tipada
	switch p := c.Parent().(type) {
	case *ast.ValueSpec:
		if p.Type != nil {
			return p.Type
		}
	default:
		// checar avô
		if vs, ok := ancestorValueSpec(c); ok && vs.Type != nil {
			return vs.Type
		}
	}
	// (3) string
	return ast.NewIdent("string")
}

func ancestorValueSpec(c *Cursor) (*ast.ValueSpec, bool) {
	// astutil não dá acesso arbitrário à cadeia; aqui mantemos simples
	if vs, ok := c.Parent().(*ast.ValueSpec); ok {
		return vs, true
	}
	return nil, false
}

// IsObfStringCall já transformado? CallExpr(Fun=?, Args=[CompositeLit([]byte{...})])
func IsObfStringCall(n ast.Node) bool {
	call, ok := n.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return false
	}
	cl, ok := call.Args[0].(*ast.CompositeLit)
	if !ok {
		return false
	}
	arr, ok := cl.Type.(*ast.ArrayType)
	if !ok {
		return false
	}
	id, ok := arr.Elt.(*ast.Ident)
	if !ok || id.Name != "byte" {
		return false
	}
	// elementos inteiros?
	for _, e := range cl.Elts {
		bl, ok := e.(*ast.BasicLit)
		if !ok || bl.Kind != token.INT {
			return false
		}
	}
	return true
}

func MakeStringFromBytes(typeExpr ast.Expr, bs []byte) ast.Expr {
	elts := make([]ast.Expr, len(bs))
	for i, b := range bs {
		elts[i] = &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(int(b))}
	}
	lit := &ast.CompositeLit{
		Type: &ast.ArrayType{Elt: ast.NewIdent("byte")},
		Elts: elts,
	}
	return &ast.CallExpr{Fun: typeExpr, Args: []ast.Expr{lit}}
}
