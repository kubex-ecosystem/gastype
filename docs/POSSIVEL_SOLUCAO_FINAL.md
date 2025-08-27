# 📌 Root Cause explicado

O seu `BoolToFlagsPass` ou `IfToBitwisePass` tá interceptando:

* **if conditions** ✅
* **atribuições** ✅

Mas não tá interceptando **uso como expressão normal** (argumento de função, return, etc).

Quando AST vê isso:

```go
fmt.Printf("%v", cfg.Debug)
```

Ele interpreta `cfg.Debug` como `*ast.SelectorExpr` **fora de um `if` ou assignment** → e como não existe lógica pra trocar nesse contexto, ele passa batido.

---

## 💡 Solução

A gente precisa de um **novo passo** (ou expandir o existente) que transforme **qualquer acesso a campo bool** mapeado → expressão bitwise.

---

### **Pass: FieldAccessToBitwisePass**

```go
package transpiler

import (
 "go/ast"
 "go/token"
)

type FieldAccessToBitwisePass struct{}

func (p *FieldAccessToBitwisePass) Name() string { return "FieldAccessToBitwise" }

func (p *FieldAccessToBitwisePass) Apply(file *ast.File, _ *token.FileSet, ctx *TranspileContext) error {
 ast.Inspect(file, func(n ast.Node) bool {
  sel, ok := n.(*ast.SelectorExpr)
  if !ok {
   return true
  }

  ident, ok := sel.X.(*ast.Ident)
  if !ok {
   return true
  }

  // Descobre struct e flag
  for structName, info := range ctx.Structs {
   if flagName, exists := info.FlagMapping[sel.Sel.Name]; exists {
    // Substituir por: (cfg.flags & FlagX) != 0
    newExpr := &ast.BinaryExpr{
     X: &ast.BinaryExpr{
      X:  &ast.SelectorExpr{X: ident, Sel: ast.NewIdent("flags")},
      Op: token.AND,
      Y:  ast.NewIdent(flagName),
     },
     Op: token.NEQ,
     Y:  ast.NewIdent("0"),
    }

    // Troca no AST
    *sel = *newExpr.(*ast.BinaryExpr).X.(*ast.SelectorExpr) // hack seguro pro AST reescrever
    return false
   }
  }

  return true
 })
 return nil
}
```

---

## 🚀 Onde encaixar no pipeline

* Rodar **depois** do `BoolToFlagsPass` e **antes** do OutputManager:

```go
--passes bool2flags,if2bitwise,assign2bitwise,field2bitwise,stringobf,jumptable
```

* Ou simplesmente incluir no `DefaultPipeline()`:

```go
case "field2bitwise":
    selected = append(selected, &FieldAccessToBitwisePass{})
```

---

## ✅ O que vai acontecer

Antes:

```go
fmt.Printf("%v %v %v", cfg.Debug, cfg.Verbose, cfg.Logging)
```

Depois:

```go
fmt.Printf("%v %v %v",
    cfg.flags&FlagConfig_Debug != 0,
    cfg.flags&FlagConfig_Verbose != 0,
    cfg.flags&FlagConfig_Logging != 0)
```
