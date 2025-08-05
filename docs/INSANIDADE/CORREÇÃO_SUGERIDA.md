# **`BoolToFlagsPass` (atualizado)**

```go
package transpiler

import (
 "fmt"
 "go/ast"
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

  if len(boolFields) == 0 {
   return true
  }

  // Nome da struct original e da nova
  structName := ts.Name.Name
  newName := structName + "Flags"

  // Cria mapping centralizado
  mapping := make(map[string]string)
  for i, field := range boolFields {
   flagName := fmt.Sprintf("Flag%s_%s", structName, strings.Title(field))
   mapping[field] = flagName
   // Guarda no contexto
   ctx.AddFlagMapping(structName, field, flagName, 1<<i)
  }

  // Registra no contexto
  ctx.AddStruct(structName, newName, boolFields)

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
 return nil
}
```

---

## **Ajuste no `TranspileContext`**

Adicionar este método:

```go
func (ctx *TranspileContext) AddFlagMapping(structName, fieldName, flagName string, bitPos int) {
 if ctx.Structs[structName] == nil {
  ctx.Structs[structName] = &StructInfo{
   OriginalName: structName,
   BoolFields:   []string{},
   FlagMapping:  map[string]string{},
  }
 }
 ctx.Structs[structName].FlagMapping[fieldName] = flagName
}
```

---

## **`IfToBitwisePass` (atualizado)**

```go
package transpiler

import (
 "go/ast"
 "go/token"
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

  // Descobre struct e flag name
  for structName, info := range ctx.Structs {
   if flagName, exists := info.FlagMapping[sel.Sel.Name]; exists {
    ifStmt.Cond = &ast.BinaryExpr{
     X: &ast.BinaryExpr{
      X:  &ast.SelectorExpr{X: ident, Sel: ast.NewIdent("flags")},
      Op: token.AND,
      Y:  ast.NewIdent(flagName),
     },
     Op: token.NEQ,
     Y:  ast.NewIdent("0"),
    }
    break
   }
  }

  return true
 })
 return nil
}
```

---

## **`AssignToBitwisePass` (atualizado)**

```go
package transpiler

import (
 "go/ast"
 "go/token"
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
  ident, ok := sel.X.(*ast.Ident)
  if !ok {
   return true
  }

  valIdent, ok := as.Rhs[0].(*ast.Ident)
  if !ok {
   return true
  }

  for structName, info := range ctx.Structs {
   if flagName, exists := info.FlagMapping[sel.Sel.Name]; exists {
    as.Lhs[0] = &ast.SelectorExpr{X: ident, Sel: ast.NewIdent("flags")}

    if valIdent.Name == "true" {
     as.Tok = token.ASSIGN
     as.Rhs[0] = &ast.BinaryExpr{
      X:  &ast.SelectorExpr{X: ident, Sel: ast.NewIdent("flags")},
      Op: token.OR_ASSIGN,
      Y:  ast.NewIdent(flagName),
     }
    } else if valIdent.Name == "false" {
     as.Tok = token.ASSIGN
     as.Rhs[0] = &ast.BinaryExpr{
      X:  &ast.SelectorExpr{X: ident, Sel: ast.NewIdent("flags")},
      Op: token.AND_NOT_ASSIGN,
      Y:  ast.NewIdent(flagName),
     }
    }
    break
   }
  }

  return true
 })
 return nil
}
```

---

💡 **O que muda com isso**

* Prefixo por struct (`FlagStruct_Field`)
* Nome único garantido
* `ctx.Structs` mantém **mapa completo**
* `IfToBitwise` e `AssignToBitwise` **consultam o mapa** — sem risco de órfãos
