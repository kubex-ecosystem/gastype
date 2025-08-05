# **1️⃣ IfToBitwisePass**

Converte:

```go
if cfg.Debug {
    println("debug on")
}
```

Para:

```go
if cfg.flags & FlagDebug != 0 {
    println("debug on")
}
```

```go
package transpiler

import (
    "go/ast"
    "strings"
)

type IfToBitwisePass struct{}

func (p *IfToBitwisePass) Name() string { return "IfToBitwise" }

func (p *IfToBitwisePass) Apply(file *ast.File, _ *token.FileSet, ctx *TranspileContext) error {
    ast.Inspect(file, func(n ast.Node) bool {
        ifStmt, ok := n.(*ast.IfStmt)
        if !ok {
            return true
        }
        sel, ok := ifStmt.Cond.(*ast.SelectorExpr)
        if !ok {
            return true
        }
        ident, ok := sel.X.(*ast.Ident)
        if !ok {
            return true
        }

        // Procura se é um campo que foi convertido pra flag
        for _, info := range ctx.Structs {
            for _, field := range info.BoolFields {
                if sel.Sel.Name == field {
                    // Substitui por bitwise check
                    ifStmt.Cond = &ast.BinaryExpr{
                        X:  &ast.SelectorExpr{X: ident, Sel: ast.NewIdent("flags")},
                        Op: token.AND,
                        Y:  ast.NewIdent("Flag" + strings.Title(field)),
                    }
                    // Compara com zero
                    ifStmt.Cond = &ast.BinaryExpr{
                        X:  ifStmt.Cond,
                        Op: token.NEQ,
                        Y:  ast.NewIdent("0"),
                    }
                }
            }
        }
        return true
    })
    return nil
}
```

---

# **2️⃣ AssignToBitwisePass**

Converte:

```go
cfg.Debug = true
cfg.Auth = false
```

Para:

```go
cfg.flags |= FlagDebug
cfg.flags &^= FlagAuth
```

```go
package transpiler

import (
    "go/ast"
    "strings"
)

type AssignToBitwisePass struct{}

func (p *AssignToBitwisePass) Name() string { return "AssignToBitwise" }

func (p *AssignToBitwisePass) Apply(file *ast.File, _ *token.FileSet, ctx *TranspileContext) error {
    ast.Inspect(file, func(n ast.Node) bool {
        as, ok := n.(*ast.AssignStmt)
        if !ok || len(as.Lhs) != 1 || len(as.Rhs) != 1 {
            return true
        }
        sel, ok := as.Lhs[0].(*ast.SelectorExpr)
        if !ok {
            return true
        }
        valIdent, ok := as.Rhs[0].(*ast.Ident)
        if !ok {
            return true
        }

        ident, ok := sel.X.(*ast.Ident)
        if !ok {
            return true
        }

        for _, info := range ctx.Structs {
            for _, field := range info.BoolFields {
                if sel.Sel.Name == field {
                    flagName := "Flag" + strings.Title(field)
                    if valIdent.Name == "true" {
                        // cfg.flags |= FlagX
                        as.Lhs[0] = &ast.SelectorExpr{X: ident, Sel: ast.NewIdent("flags")}
                        as.Tok = token.ASSIGN
                        as.Rhs[0] = &ast.BinaryExpr{
                            X:  &ast.SelectorExpr{X: ident, Sel: ast.NewIdent("flags")},
                            Op: token.OR_ASSIGN,
                            Y:  ast.NewIdent(flagName),
                        }
                    } else if valIdent.Name == "false" {
                        // cfg.flags &^= FlagX
                        as.Lhs[0] = &ast.SelectorExpr{X: ident, Sel: ast.NewIdent("flags")}
                        as.Tok = token.ASSIGN
                        as.Rhs[0] = &ast.BinaryExpr{
                            X:  &ast.SelectorExpr{X: ident, Sel: ast.NewIdent("flags")},
                            Op: token.AND_NOT_ASSIGN,
                            Y:  ast.NewIdent(flagName),
                        }
                    }
                }
            }
        }
        return true
    })
    return nil
}
```

---

# **3️⃣ Como plugar no Engine**

No teu `transpileCmd`:

```go
for _, p := range strings.Split(passes, ",") {
    switch p {
    case "bool2flags":
        engine.Passes = append(engine.Passes, &transpiler.BoolToFlagsPass{})
    case "if2bitwise":
        engine.Passes = append(engine.Passes, &transpiler.IfToBitwisePass{})
    case "assign2bitwise":
        engine.Passes = append(engine.Passes, &transpiler.AssignToBitwisePass{})
    }
}
```

---

# **4️⃣ Rodando multi-pass**

```bash
# Rodar todos de uma vez
gastype transpile ./ \
  --passes bool2flags,if2bitwise,assign2bitwise \
  --map ./context.map.json

# Só ver o que seria feito
gastype transpile ./ \
  --passes bool2flags,if2bitwise,assign2bitwise \
  --dry-run
```
