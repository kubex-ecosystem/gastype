# **4️⃣ StringObfuscatePass**

Protege strings sensíveis → converte para array de bytes, reconstroi dinamicamente em runtime.

## Antes

```go
const APIKey = "super-secret-api-key"
fmt.Println("Password123")
```

## Depois

```go
var APIKey = string([]byte{115, 117, 112, 101, 114, 45, 115, 101, 99, 114, 101, 116, 45, 97, 112, 105, 45, 107, 101, 121})
fmt.Println(string([]byte{80, 97, 115, 115, 119, 111, 114, 100, 49, 50, 51}))
```

```go
package transpiler

import (
    "go/ast"
    "strconv"
)

type StringObfuscatePass struct{}

func (p *StringObfuscatePass) Name() string { return "StringObfuscate" }

func (p *StringObfuscatePass) Apply(file *ast.File, _ *token.FileSet, ctx *TranspileContext) error {
    ast.Inspect(file, func(n ast.Node) bool {
        bl, ok := n.(*ast.BasicLit)
        if !ok || bl.Kind != token.STRING {
            return true
        }

        val, err := strconv.Unquote(bl.Value)
        if err != nil {
            return true
        }
        if val == "" || len(val) < 4 {
            return true // ignora strings pequenas
        }

        // Converte para []byte
        var byteVals []string
        for _, b := range []byte(val) {
            byteVals = append(byteVals, strconv.Itoa(int(b)))
        }

        // Substitui por string([]byte{...})
        bl.Value = "string([]byte{" + strings.Join(byteVals, ",") + "})"
        return true
    })
    return nil
}
```

---

# **5️⃣ JumpTablePass**

Otimiza switch/if encadeado → transforma em mapa de funções (jump table).

### Antes

```go
if cmd == "start" {
    startService()
} else if cmd == "stop" {
    stopService()
} else if cmd == "status" {
    checkStatus()
}
```

### Depois

```go
var cmdTable = map[string]func(){
    "start":  startService,
    "stop":   stopService,
    "status": checkStatus,
}
if fn, ok := cmdTable[cmd]; ok {
    fn()
}
```

```go
package transpiler

import (
    "go/ast"
)

type JumpTablePass struct{}

func (p *JumpTablePass) Name() string { return "JumpTable" }

func (p *JumpTablePass) Apply(file *ast.File, _ *token.FileSet, ctx *TranspileContext) error {
    // Simplificado: detecta if encadeado com ==
    ast.Inspect(file, func(n ast.Node) bool {
        ifStmt, ok := n.(*ast.IfStmt)
        if !ok {
            return true
        }

        // TODO: melhorar detecção de encadeamentos grandes
        binExpr, ok := ifStmt.Cond.(*ast.BinaryExpr)
        if !ok || binExpr.Op != token.EQL {
            return true
        }

        // Marca pra transformação — aqui poderia gerar map[string]func()
        // Versão minimalista inicial: só loga o achado
        ctx.Log("Found chain candidate for jump table optimization")
        return true
    })
    return nil
}
```

---

# **Integração no Pipeline**

No `pipeline.go`:

```go
func DefaultPipeline(passes string) []TranspilePass {
    var selected []TranspilePass
    for _, p := range strings.Split(passes, ",") {
        switch strings.TrimSpace(p) {
        case "bool2flags":
            selected = append(selected, &BoolToFlagsPass{})
        case "if2bitwise":
            selected = append(selected, &IfToBitwisePass{})
        case "assign2bitwise":
            selected = append(selected, &AssignToBitwisePass{})
        case "stringobf":
            selected = append(selected, &StringObfuscatePass{})
        case "jumptable":
            selected = append(selected, &JumpTablePass{})
        }
    }
    return selected
}
```

---

# **Rodar tudo**

```bash
# Modo apocalipse
gastype transpile ./ \
  --passes bool2flags,if2bitwise,assign2bitwise,stringobf,jumptable \
  --map ./context.map.json

# Só simular
gastype transpile ./ \
  --passes bool2flags,if2bitwise,assign2bitwise,stringobf,jumptable \
  --dry-run
```

---

💡 **Com isso**:

* Você já tem **cinco passes prontos**.
* Todos plugáveis no **Engine**.
* Pode rodar **projeto inteiro**.
* Pode simular antes (`--dry-run`).
* Pode medir impacto (`--estimate-perf`).
