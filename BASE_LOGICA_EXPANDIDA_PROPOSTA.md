
# Versão Expandida — Transpiler AST com "Reescrita Completa" (tentativa)

```go
package cmd

import (
    "fmt"
    "go/ast"
    "go/parser"
    "go/printer"
    "go/token"
    "os"
    "path/filepath"
    "strings"
)

func TranspileBoolToFlagsFull(inputFile, outputDir string) error {
    fset := token.NewFileSet()
    file, err := parser.ParseFile(fset, inputFile, nil, parser.ParseComments)
    if err != nil {
        return err
    }

    structBools := map[string][]string{} // StructName -> []boolField

    // 1️⃣ Detecta structs com bool
    ast.Inspect(file, func(n ast.Node) bool {
        ts, ok := n.(*ast.TypeSpec)
        if !ok {
            return true
        }
        st, ok := ts.Type.(*ast.StructType)
        if !ok {
            return true
        }

        for _, field := range st.Fields.List {
            if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "bool" {
                for _, name := range field.Names {
                    structBools[ts.Name.Name] = append(structBools[ts.Name.Name], name.Name)
                }
            }
        }
        return true
    })

    // 2️⃣ Reescreve structs → flags + const
    for structName, bools := range structBools {
        fmt.Printf("[INFO] Struct %s → %d bool(s) → flags\n", structName, len(bools))

        // Constantes
        constDecl := &ast.GenDecl{Tok: token.CONST}
        for i, bf := range bools {
            name := "Flag" + strings.Title(bf)
            constDecl.Specs = append(constDecl.Specs, &ast.ValueSpec{
                Names:  []*ast.Ident{ast.NewIdent(name)},
                Type:   ast.NewIdent(structName + "Flags"),
                Values: []ast.Expr{&ast.BinaryExpr{
                    X:  &ast.BasicLit{Kind: token.INT, Value: "1"},
                    Op: token.SHL,
                    Y:  &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", i)},
                }},
            })
        }
        file.Decls = append([]ast.Decl{constDecl}, file.Decls...)

        // Renomeia struct + substitui campos
        ast.Inspect(file, func(n ast.Node) bool {
            ts, ok := n.(*ast.TypeSpec)
            if !ok {
                return true
            }
            if ts.Name.Name == structName {
                ts.Name.Name = structName + "Flags"
                st := ts.Type.(*ast.StructType)
                st.Fields.List = []*ast.Field{
                    {
                        Names: []*ast.Ident{ast.NewIdent("flags")},
                        Type:  ast.NewIdent("uint64"),
                    },
                }
            }
            return true
        })
    }

    // 3️⃣ Substitui acessos `.BoolField` → bitwise
    ast.Inspect(file, func(n ast.Node) bool {
        sel, ok := n.(*ast.SelectorExpr)
        if !ok {
            return true
        }
        ident, ok := sel.X.(*ast.Ident)
        if !ok {
            return true
        }

        for structName, fields := range structBools {
            for _, field := range fields {
                if sel.Sel.Name == field {
                    // Envolve em flags & mask
                    sel.Sel.Name = "flags & Flag" + strings.Title(field) + " != 0"
                    ident.Name = ident.Name // mantém
                }
            }
        }
        return true
    })

    // 4️⃣ Substitui atribuições `.BoolField = true/false`
    ast.Inspect(file, func(n ast.Node) bool {
        as, ok := n.(*ast.AssignStmt)
        if !ok {
            return true
        }

        if len(as.Lhs) != 1 || len(as.Rhs) != 1 {
            return true
        }

        sel, ok := as.Lhs[0].(*ast.SelectorExpr)
        if !ok {
            return true
        }
        val, ok := as.Rhs[0].(*ast.Ident)
        if !ok {
            return true
        }

        for _, fields := range structBools {
            for _, field := range fields {
                if sel.Sel.Name == field {
                    if val.Name == "true" {
                        // vira flags |= mask
                        as.Lhs[0] = &ast.SelectorExpr{
                            X:   sel.X,
                            Sel: ast.NewIdent("flags"),
                        }
                        as.Tok = token.ASSIGN
                        as.Rhs[0] = &ast.BinaryExpr{
                            X:  &ast.SelectorExpr{X: sel.X, Sel: ast.NewIdent("flags")},
                            Op: token.OR_ASSIGN,
                            Y:  ast.NewIdent("Flag" + strings.Title(field)),
                        }
                    } else if val.Name == "false" {
                        // vira flags &^= mask
                        as.Lhs[0] = &ast.SelectorExpr{
                            X:   sel.X,
                            Sel: ast.NewIdent("flags"),
                        }
                        as.Tok = token.ASSIGN
                        as.Rhs[0] = &ast.BinaryExpr{
                            X:  &ast.SelectorExpr{X: sel.X, Sel: ast.NewIdent("flags")},
                            Op: token.AND_NOT_ASSIGN,
                            Y:  ast.NewIdent("Flag" + strings.Title(field)),
                        }
                    }
                }
            }
        }
        return true
    })

    // Salva saída
    os.MkdirAll(outputDir, 0755)
    outputFile := filepath.Join(outputDir, filepath.Base(inputFile))
    out, err := os.Create(outputFile)
    if err != nil {
        return err
    }
    defer out.Close()

    printer.Fprint(out, fset, file)
    fmt.Printf("[OK] Arquivo convertido salvo em: %s\n", outputFile)
    return nil
}
```
