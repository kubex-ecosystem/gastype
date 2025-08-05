# 🎯 **MVP do Transpile Real**

## 1️⃣ Entrada (original legível)

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

---

### 2️⃣ Saída esperada (bitwise legível)

```go
package main

type ServiceConfigFlags uint8

const (
    FlagDebug ServiceConfigFlags = 1 << iota
    FlagAuth
    FlagCache
)

func main() {
    var cfg ServiceConfigFlags
    cfg |= FlagDebug
    cfg |= FlagCache

    if cfg&FlagDebug != 0 {
        println("Debug ativo")
    }
}
```

---

## ⚡ Passo rápido pra ver isso **hoje** no GasType

No `cmd/transpile.go`:

```go
package cmd

import (
    "fmt"
    "go/ast"
    "go/parser"
    "go/printer"
    "go/token"
    "os"
    "strings"
)

func TranspileBoolToFlags(inputFile, outputFile string) error {
    fset := token.NewFileSet()
    f, err := parser.ParseFile(fset, inputFile, nil, parser.ParseComments)
    if err != nil {
        return err
    }

    // Procura structs com bool
    ast.Inspect(f, func(n ast.Node) bool {
        ts, ok := n.(*ast.TypeSpec)
        if !ok {
            return true
        }
        st, ok := ts.Type.(*ast.StructType)
        if !ok {
            return true
        }

        var hasBool bool
        for _, field := range st.Fields.List {
            if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "bool" {
                hasBool = true
            }
        }

        if hasBool {
            fmt.Printf("[INFO] Struct %s tem bool → converter para flags\n", ts.Name)
            ts.Name.Name = ts.Name.Name + "Flags"
            st.Fields.List = []*ast.Field{} // limpa campos
        }
        return true
    })

    out, err := os.Create(outputFile)
    if err != nil {
        return err
    }
    defer out.Close()

    printer.Fprint(out, fset, f)
    return nil
}
```

---

## 🚀 Rodar agora

```bash
go run ./cmd main.go ./out_optimized/main.go
```
