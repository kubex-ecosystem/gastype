// Package pass provides StringObfuscatePass for obfuscating string literals
package pass

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

	stdastutil "golang.org/x/tools/go/ast/astutil"

	"github.com/rafa-mori/gastype/internal/astutil"
)

// StringObfuscatePass converts string literals to byte array reconstructions
// Example: "password123" → string([]byte{112,97,115,115,119,111,114,100,49,50,51})
type StringObfuscatePass struct{}

func NewStringObfuscatePass() *StringObfuscatePass {
	return &StringObfuscatePass{}
}

func (p *StringObfuscatePass) Name() string {
	return "StringObfuscate"
}

func (p *StringObfuscatePass) Apply(file *ast.File, fset *token.FileSet, ctx *astutil.TranspileContext) error {
	transformations := 0

	// ========== FASE 1: Identificação ==========
	// Mapeia literais que não devem ser obfuscadas
	skip := make(map[*ast.BasicLit]bool)

	// Import paths
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			for _, spec := range genDecl.Specs {
				if importSpec, ok := spec.(*ast.ImportSpec); ok && importSpec.Path != nil {
					if importSpec.Path.ValuePos.IsValid() && importSpec.Path.Value != "" {
						basicLit := importSpec.Path
						if basicLit != nil && basicLit.Kind == token.STRING {
							skip[basicLit] = true
						}
					}
				}
			}
		}
	}

	// Const strings
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.CONST {
			ast.Inspect(genDecl, func(n ast.Node) bool {
				if bl, ok := n.(*ast.BasicLit); ok && bl.Kind == token.STRING {
					skip[bl] = true
				}
				return true
			})
		}
	}

	// Struct tags
	ast.Inspect(file, func(n ast.Node) bool {
		if field, ok := n.(*ast.Field); ok && field.Tag != nil {
			fdTag := field.Tag.Value
			if fdTag != "" {
				skip[field.Tag] = true
			}
		}
		return true
	})

	// ========== FASE 2: Substituição ==========
	stdastutil.Apply(file, func(cr *stdastutil.Cursor) bool {
		bl, ok := cr.Node().(*ast.BasicLit)
		if !ok || bl.Kind != token.STRING {
			return true
		}

		// Skip casos protegidos
		if skip[bl] {
			return true
		}

		// Valor puro
		val, err := strconv.Unquote(bl.Value)
		if err != nil {
			return true
		}

		// Regras de exclusão adicionais
		if val == "" || len(val) < 4 {
			return true
		}
		commonWords := []string{"main", "func", "package", "import", "var", "const", "if", "else", "for", "range"}
		for _, word := range commonWords {
			if val == word {
				return true
			}
		}

		// Monta byte array
		var byteVals []string
		for _, b := range []byte(val) {
			byteVals = append(byteVals, strconv.Itoa(int(b)))
		}

		// Novo nó AST → string([]byte{...})
		obfuscated := &ast.CallExpr{
			Fun: ast.NewIdent("string"),
			Args: []ast.Expr{
				&ast.CompositeLit{
					Type: &ast.ArrayType{
						Elt: ast.NewIdent("byte"),
					},
					Elts: func() []ast.Expr {
						var elts []ast.Expr
						for _, bVal := range byteVals {
							elts = append(elts, &ast.BasicLit{
								Kind:  token.INT,
								Value: bVal,
							})
						}
						return elts
					}(),
				},
			},
		}

		// Substitui no AST
		cr.Replace(obfuscated)

		transformations++
		fmt.Printf("    🔒 Obfuscated string literal: %q → byte array\n", val)

		return true
	}, nil)

	if transformations > 0 {
		fmt.Printf("  🔄 StringObfuscatePass: %d transformations applied\n", transformations)
	}

	return nil

}
