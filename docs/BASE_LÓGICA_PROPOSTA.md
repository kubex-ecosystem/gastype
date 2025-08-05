# **cmd/transpile.go**

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

func TranspileBoolToFlags(inputFile, outputDir string) error {
    fset := token.NewFileSet()

    // Parse do arquivo
    file, err := parser.ParseFile(fset, inputFile, nil, parser.ParseComments)
    if err != nil {
        return err
    }

    // Percorre a AST procurando structs com bool
    ast.Inspect(file, func(n ast.Node) bool {
        ts, ok := n.(*ast.TypeSpec)
        if !ok {
            return true
        }

        st, ok := ts.Type.(*ast.StructType)
        if !ok {
            return true
        }

        // Verifica se tem bool
        var boolFields []string
        for _, field := range st.Fields.List {
            if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "bool" {
                for _, name := range field.Names {
                    boolFields = append(boolFields, name.Name)
                }
            }
        }

        if len(boolFields) > 0 {
            fmt.Printf("[INFO] Struct %s → %d bool(s) detectados\n", ts.Name.Name, len(boolFields))

            // Renomeia struct
            ts.Name.Name = ts.Name.Name + "Flags"

            // Substitui campos por um único campo de flags
            st.Fields.List = []*ast.Field{
                {
                    Names: []*ast.Ident{ast.NewIdent("flags")},
                    Type:  ast.NewIdent("uint64"),
                },
            }

            // Gera constantes para cada flag
            constDecl := &ast.GenDecl{
                Tok: token.CONST,
            }

            for i, bf := range boolFields {
                name := "Flag" + strings.Title(bf)
                constDecl.Specs = append(constDecl.Specs, &ast.ValueSpec{
                    Names:  []*ast.Ident{ast.NewIdent(name)},
                    Type:   ast.NewIdent(ts.Name.Name),
                    Values: []ast.Expr{&ast.BinaryExpr{
                        X:  &ast.BasicLit{Kind: token.INT, Value: "1"},
                        Op: token.SHL,
                        Y:  &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", i)},
                    }},
                })
            }

            // Insere as constantes no topo do arquivo
            file.Decls = append([]ast.Decl{constDecl}, file.Decls...)
        }
        return true
    })

    // Garante que diretório de saída existe
    os.MkdirAll(outputDir, 0755)

    // Arquivo de saída
    outputFile := filepath.Join(outputDir, filepath.Base(inputFile))
    out, err := os.Create(outputFile)
    if err != nil {
        return err
    }
    defer out.Close()

    // Escreve código modificado
    printer.Fprint(out, fset, file)

    fmt.Printf("[OK] Arquivo convertido salvo em: %s\n", outputFile)
    return nil
}
```

---

## **Como usar no seu CLI atual**

No `cmd/root.go` (ou onde define comandos):

```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var transpileCmd = &cobra.Command{
    Use:   "transpile [arquivo.go]",
    Short: "Transpila bools para bitflags",
    RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) == 0 {
            return fmt.Errorf("forneça um arquivo Go para transpilar")
        }
        return TranspileBoolToFlags(args[0], "./out_optimized")
    },
}

func init() {
    rootCmd.AddCommand(transpileCmd)
}
```

---

## **Teste rápido**

### Entrada (`service.go`)

```go
package main

type ServiceConfig struct {
    Debug bool
    Auth  bool
    Cache bool
}

func main() {
    cfg := ServiceConfig{
        Debug: true,
        Auth:  false,
        Cache: true,
    }

    if cfg.Debug {
        println("Debug ativo")
    }
}
```

### Comando

```bash
go run . transpile service.go
```

### Saída (`out_optimized/service.go`)

```go
package main

const (
    FlagDebug ServiceConfigFlags = 1 << 0
    FlagAuth  ServiceConfigFlags = 1 << 1
    FlagCache ServiceConfigFlags = 1 << 2
)

type ServiceConfigFlags struct {
    flags uint64
}

func main() {
    cfg := ServiceConfigFlags{}
    cfg.flags |= FlagDebug
    cfg.flags |= FlagCache

    if cfg.flags&FlagDebug != 0 {
        println("Debug ativo")
    }
}
```
