# **1️⃣ Estrutura Base — `engine.go`**

```go
package transpiler

import (
    "fmt"
    "go/ast"
    "go/parser"
    "go/printer"
    "go/token"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
)

// Engine coordena passes e contexto
type Engine struct {
    Ctx    *TranspileContext
    Passes []TranspilePass
}

// Interface para qualquer transformação AST
type TranspilePass interface {
    Name() string
    Apply(file *ast.File, fset *token.FileSet, ctx *TranspileContext) error
}

// Descobre todos arquivos Go, recursivamente
func DiscoverGoFiles(root string) ([]string, error) {
    var files []string
    err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if !d.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
            files = append(files, path)
        }
        return nil
    })
    return files, err
}

// Executa a engine
func (e *Engine) Run(root string) error {
    files, err := DiscoverGoFiles(root)
    if err != nil {
        return err
    }

    for _, filePath := range files {
        fmt.Printf("🔍 Processing %s\n", filePath)
        fset := token.NewFileSet()
        astFile, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
        if err != nil {
            return err
        }

        for _, pass := range e.Passes {
            fmt.Printf("  ⚙️ Applying pass: %s\n", pass.Name())
            if err := pass.Apply(astFile, fset, e.Ctx); err != nil {
                return err
            }
        }

        if !e.Ctx.DryRun {
            outPath := filepath.Join(e.Ctx.OutputDir, filepath.Base(filePath))
            os.MkdirAll(filepath.Dir(outPath), 0755)
            outFile, _ := os.Create(outPath)
            printer.Fprint(outFile, fset, astFile)
            outFile.Close()
        }
    }

    if e.Ctx.MapFile != "" {
        return e.Ctx.SaveMap()
    }
    return nil
}
```

---

## **2️⃣ Exemplo de Pass — `bool_to_flags.go`**

```go
package transpiler

import (
    "fmt"
    "go/ast"
    "go/token"
    "strings"
)

type BoolToFlagsPass struct{}

func (p *BoolToFlagsPass) Name() string { return "BoolToFlags" }

func (p *BoolToFlagsPass) Apply(file *ast.File, fset *token.FileSet, ctx *TranspileContext) error {
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

        if len(boolFields) > 0 {
            newName := ts.Name.Name + "Flags"
            ctx.AddStruct(ts.Name.Name, newName, boolFields)
            fmt.Printf("  📝 Converting struct %s → %s (%d bool fields)\n",
                ts.Name.Name, newName, len(boolFields))
            ts.Name.Name = newName
            st.Fields.List = []*ast.Field{
                {Names: []*ast.Ident{ast.NewIdent("flags")}, Type: ast.NewIdent("uint64")},
            }
        }
        return true
    })
    return nil
}
```

---

## **3️⃣ CLI — `cmd/cli/transpile.go`**

```go
package cli

import (
    "fmt"
    "strings"

    "github.com/spf13/cobra"
    "seu/projeto/transpiler"
)

var dryRun bool
var estimate bool
var passes string
var mapFile string
var outputDir string

func init() {
    transpileCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Apenas analisar, sem salvar")
    transpileCmd.Flags().BoolVar(&estimate, "estimate-perf", false, "Estimar ganhos de performance")
    transpileCmd.Flags().StringVar(&passes, "passes", "bool2flags", "Lista de passes a aplicar")
    transpileCmd.Flags().StringVar(&mapFile, "map", "", "Gera arquivo de mapeamento JSON")
    transpileCmd.Flags().StringVar(&outputDir, "o", "./out_optimized", "Diretório de saída")
}

var transpileCmd = &cobra.Command{
    Use:   "transpile [pasta|arquivo]",
    Short: "Transpila projeto Go para otimizações bitwise",
    RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) == 0 {
            return fmt.Errorf("Forneça um diretório ou arquivo Go")
        }

        ctx := transpiler.NewContext(args[0], outputDir, false, mapFile)
        ctx.DryRun = dryRun

        engine := &transpiler.Engine{Ctx: ctx}

        for _, p := range strings.Split(passes, ",") {
            switch p {
            case "bool2flags":
                engine.Passes = append(engine.Passes, &transpiler.BoolToFlagsPass{})
            }
        }

        if err := engine.Run(args[0]); err != nil {
            return err
        }
        if estimate {
            ctx.EstimatePerformance()
        }
        return nil
    },
}
```

---

## **4️⃣ Uso**

```bash
# Rodar no projeto inteiro
gastype transpile ./ --map ./context.map.json --passes bool2flags

# Só simular
gastype transpile ./ --dry-run --passes bool2flags

# Estimar performance antes de aplicar
gastype transpile ./ --estimate-perf --dry-run
```
