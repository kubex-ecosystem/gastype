# FRONTEIRA FINAL: stdlib sendo reescrita

O AST do Go tá considerando pacotes da stdlib como imports locais, e o `rewriteImports()` tá reescrevendo eles com o `modulePath` na frente.
Exemplo do erro abaixo.

```go
 "github.com/rafa-mori/gobe/encoding/json"
 "fmt"
 "io"
 "github.com/rafa-mori/gobe/net/http"
 "strconv"
 "strings"
 "time"

 manifest "github.com/rafa-mori/gobe/info"
 "github.com/rafa-mori/gobe/logger"
 "github.com/spf13/cobra"
```

## 📌 Por que tá acontecendo

O `rewriteImports()` que coloquei no **OutputManager** tá com a lógica:

```go
if !strings.Contains(path, ".") { // se não tem ponto, trata como local
    imp.Path.Value = strconv.Quote(modulePath + "/" + path)
}
```

Só que o AST considera `encoding/json` e `net/http` como **path sem ponto** — porque eles são do Go standard library.
A função acha que é “import local” e enfia seu `modulePath` na frente → `"github.com/rafa-mori/gobe/encoding/json"`.
💀 Resultado: **stdlib sequestrada pro seu repositório**.

---

## 🎯 A correção

Temos que **detectar pacotes da stdlib** e **pular rewrite** nesses casos.

---

### **Forma mais segura**

* Se **primeiro segmento** NÃO contém ponto e NÃO contém barra `/` que bata com estrutura local, **é stdlib**.
* Ou consultar a **Go List** (`go list std`) e montar um set com todos pacotes stdlib — mais preciso.

---

### **Patch rápido**

Substituir:

```go
if strings.Contains(path, ".") {
    continue
}
```

Por:

```go
// Detecta se é stdlib: sem ponto, mas não é relativo ao projeto
if isStdLib(path) {
    continue
}
```

E criar:

```go
func isStdLib(importPath string) bool {
    // Caminhos como "fmt", "io", "time", "encoding/json", "net/http"...
    if !strings.Contains(importPath, ".") {
        return true
    }
    return false
}
```

---

### **Exemplo corrigido**

Antes:

```go
"encoding/json" → "github.com/rafa-mori/gobe/encoding/json"
"net/http"      → "github.com/rafa-mori/gobe/net/http"
```

Depois:

```go
"encoding/json" // mantido
"net/http"      // mantido
```

---

📌 **Extra**: se quiser deixar **100% blindado**, eu monto uma função que:

* Executa `go list std` e preenche um map.
* Só faz rewrite se NÃO estiver nesse set.
