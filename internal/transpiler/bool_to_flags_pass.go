// Package transpiler provides pass-based AST transformations
package transpiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// BoolToFlagsPass transforms boolean struct fields to bitwise flags
type BoolToFlagsPass struct{}

// Name returns the name of this pass
func (p *BoolToFlagsPass) Name() string {
	return "BoolToFlags"
}

// Apply executa a transformação bool-to-flags completa
func (p *BoolToFlagsPass) Apply(file *ast.File, fset *token.FileSet, ctx *TranspileContext) error {
	// Fase 1: Transformar structs
	p.transformStructs(file, ctx)

	// Fase 2: Adicionar constantes
	p.addFlagConstants(file, ctx)

	// Fase 3: Transformar struct literals
	p.transformStructLiterals(file, ctx)

	// Fase 4: Transformar field access
	p.transformFieldAccess(file, ctx)

	return nil
} // transformStructs encontra structs com bool fields e os transforma
func (p *BoolToFlagsPass) transformStructs(file *ast.File, ctx *TranspileContext) {
	ast.Inspect(file, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			return true
		}

		var boolFields []string
		for _, field := range st.Fields.List {
			if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "bool" {
				for _, name := range field.Names {
					boolFields = append(boolFields, name.Name)
				}
			}
		}

		if len(boolFields) == 0 {
			return true
		}

		// Nome da struct original e da nova
		structName := ts.Name.Name
		newName := structName + "Flags"

		// Cria mapping centralizado
		for i, field := range boolFields {
			flagName := fmt.Sprintf("Flag%s_%s", structName, strings.Title(field))
			ctx.AddFlagMapping(structName, field, flagName, 1<<i)
		}

		// Registra no contexto
		ctx.AddStruct("", structName, newName, boolFields)

		// Troca struct para usar flags
		ts.Name.Name = newName
		st.Fields.List = []*ast.Field{
			{
				Names: []*ast.Ident{ast.NewIdent("flags")},
				Type:  ast.NewIdent("uint64"),
			},
		}

		return true
	})
}

// addFlagConstants adiciona as constantes das flags no arquivo
func (p *BoolToFlagsPass) addFlagConstants(file *ast.File, ctx *TranspileContext) {
	for structName, structInfo := range ctx.Structs {
		if len(structInfo.BoolFields) == 0 {
			continue
		}

		specs := []ast.Spec{}
		for i, field := range structInfo.BoolFields {
			flagName := fmt.Sprintf("Flag%s_%s", structName, strings.Title(field))

			spec := &ast.ValueSpec{
				Names: []*ast.Ident{ast.NewIdent(flagName)},
				Type:  ast.NewIdent("uint64"),
				Values: []ast.Expr{
					&ast.BinaryExpr{
						X: &ast.BasicLit{
							Kind:  token.INT,
							Value: "1",
						},
						Op: token.SHL,
						Y: &ast.BasicLit{
							Kind:  token.INT,
							Value: fmt.Sprintf("%d", i),
						},
					},
				},
			}
			specs = append(specs, spec)
		}

		// Criar declaração const
		constDecl := &ast.GenDecl{
			Tok:   token.CONST,
			Specs: specs,
		}

		// Inserir após imports
		insertPos := 0
		for i, decl := range file.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
				insertPos = i + 1
			}
		}

		// Inserir a declaração const
		newDecls := make([]ast.Decl, 0, len(file.Decls)+1)
		newDecls = append(newDecls, file.Decls[:insertPos]...)
		newDecls = append(newDecls, constDecl)
		newDecls = append(newDecls, file.Decls[insertPos:]...)
		file.Decls = newDecls
	}
}

// transformStructLiterals transforma inicialização de structs
func (p *BoolToFlagsPass) transformStructLiterals(file *ast.File, ctx *TranspileContext) {
	ast.Inspect(file, func(n ast.Node) bool {
		cl, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		// Verifica se é uma struct que foi transformada
		if ident, ok := cl.Type.(*ast.Ident); ok {
			for originalName, structInfo := range ctx.Structs {
				if ident.Name == originalName {
					// Mudar o tipo para a nova struct
					cl.Type = ast.NewIdent(structInfo.NewName)

					// Converter elementos para flags
					var flagsValue uint64
					for _, elt := range cl.Elts {
						if kv, ok := elt.(*ast.KeyValueExpr); ok {
							if key, ok := kv.Key.(*ast.Ident); ok {
								if val, ok := kv.Value.(*ast.Ident); ok && val.Name == "true" {
									// Encontrar o bit position para este field
									for i, field := range structInfo.BoolFields {
										if field == key.Name {
											flagsValue |= 1 << i
											break
										}
									}
								}
							}
						}
					}

					// Substituir por inicialização flags
					cl.Elts = []ast.Expr{
						&ast.KeyValueExpr{
							Key: ast.NewIdent("flags"),
							Value: &ast.BasicLit{
								Kind:  token.INT,
								Value: fmt.Sprintf("%d", flagsValue),
							},
						},
					}
					return false
				}
			}
		}
		return true
	})
}

// transformFieldAccess transforma acessos a campos bool transformados
func (p *BoolToFlagsPass) transformFieldAccess(file *ast.File, ctx *TranspileContext) {
	ast.Inspect(file, func(n ast.Node) bool {
		// Não implementamos ainda - deixar IfToBitwisePass handle field access
		return true
	})
}
