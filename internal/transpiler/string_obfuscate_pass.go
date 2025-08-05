// Package transpiler provides StringObfuscatePass for obfuscating string literals
package transpiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

// StringObfuscatePass converts string literals to byte array reconstructions
// Transforms: "password123" → string([]byte{112, 97, 115, 115, 119, 111, 114, 100, 49, 50, 51})
type StringObfuscatePass struct{}

func (p *StringObfuscatePass) Name() string {
	return "StringObfuscate"
}

func (p *StringObfuscatePass) Apply(file *ast.File, fset *token.FileSet, ctx *TranspileContext) error {
	transformations := 0

	// Track import declarations to skip obfuscating import paths
	importSpecs := make(map[*ast.BasicLit]bool)
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			for _, spec := range genDecl.Specs {
				if importSpec, ok := spec.(*ast.ImportSpec); ok && importSpec.Path != nil {
					importSpecs[importSpec.Path] = true
				}
			}
		}
	}

	ast.Inspect(file, func(n ast.Node) bool {
		bl, ok := n.(*ast.BasicLit)
		if !ok || bl.Kind != token.STRING {
			return true
		}

		// Skip import paths completely
		if importSpecs[bl] {
			return true
		}

		// Unquote the string to get actual value
		val, err := strconv.Unquote(bl.Value)
		if err != nil {
			return true
		}

		// Skip small strings or very common ones
		if val == "" || len(val) < 4 {
			return true
		}

		// Skip very common words that would look suspicious if obfuscated
		commonWords := []string{"main", "func", "package", "import", "var", "const", "if", "else", "for", "range"}
		for _, word := range commonWords {
			if val == word {
				return true
			}
		}

		// Convert string to byte array
		var byteVals []string
		for _, b := range []byte(val) {
			byteVals = append(byteVals, strconv.Itoa(int(b)))
		}

		// Create the obfuscated expression: string([]byte{...})
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

		// Replace the original string with obfuscated version
		// We need to replace in the parent node, but this is a complex operation
		// For now, we'll modify the Value directly as a string representation
		bl.Value = fmt.Sprintf("string([]byte{%s})", strings.Join(byteVals, ", "))
		bl.Kind = token.STRING // Keep as string but with obfuscated content

		transformations++
		fmt.Printf("    🔒 Obfuscated string literal: %q → byte array\n", val)

		_ = obfuscated // Suppress unused variable warning for now

		return true
	})

	if transformations > 0 {
		fmt.Printf("  🔄 StringObfuscatePass: %d transformations applied\n", transformations)
	}

	return nil
}
