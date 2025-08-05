# **🚀 GasType Engine — Pipeline Consolidado**

**Todos os 5 passes já integrados:**

* `bool2flags` → Structs com bool → flags bitwise.
* `if2bitwise` → Condições com bool → checks bitwise.
* `assign2bitwise` → Atribuições bool → set/clear bitwise.
* `stringobf` → Strings sensíveis → array de bytes.
* `jumptable` → Otimiza if encadeado → jump table.

---

## **`pipeline.go`**

```go
package transpiler

import "strings"

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

## **`cmd/transpile.go`**

```go
package cmd

import (
    "fmt"
    "strings"

    "github.com/spf13/cobra"
    "seu/projeto/transpiler"
)

var (
    dryRun    bool
    estimate  bool
    passes    string
    mapFile   string
    outputDir string
)

func init() {
    transpileCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Apenas analisar, sem salvar arquivos")
    transpileCmd.Flags().BoolVar(&estimate, "estimate-perf", false, "Estimar ganhos de performance")
    transpileCmd.Flags().StringVar(&passes, "passes", "bool2flags,if2bitwise,assign2bitwise,stringobf,jumptable", "Lista de passes")
    transpileCmd.Flags().StringVar(&mapFile, "map", "./context.map.json", "Gera arquivo de mapeamento JSON")
    transpileCmd.Flags().StringVar(&outputDir, "o", "./out_optimized", "Diretório de saída")
}

var transpileCmd = &cobra.Command{
    Use:   "transpile [pasta|arquivo]",
    Short: "Transpila projeto Go para otimizações bitwise e obfuscação",
    RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) == 0 {
            return fmt.Errorf("Forneça um diretório ou arquivo Go")
        }

        ctx := transpiler.NewContext(args[0], outputDir, false, mapFile)
        ctx.DryRun = dryRun

        engine := &transpiler.Engine{
            Ctx:    ctx,
            Passes: transpiler.DefaultPipeline(passes),
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

## **Rodando o “Modo Apocalipse”**

```bash
# Roda tudo em um projeto inteiro (GoBE, MCP, Kubex...)
gastype transpile ./ \
  --passes bool2flags,if2bitwise,assign2bitwise,stringobf,jumptable \
  --map ./context.map.json

# Só simular antes
gastype transpile ./ \
  --passes bool2flags,if2bitwise,assign2bitwise,stringobf,jumptable \
  --dry-run

# Estimar ganhos antes de aplicar
gastype transpile ./ \
  --passes bool2flags,if2bitwise,assign2bitwise,stringobf,jumptable \
  --estimate-perf --dry-run
```

---

## **Saída esperada (exemplo)**

```
🔍 Processing ./internal/config/service.go
  ⚙️ Applying pass: BoolToFlags
    📝 Converting struct ServiceConfig → ServiceConfigFlags (3 bool fields)
  ⚙️ Applying pass: IfToBitwise
    ⚡ Transformed if cfg.Debug → bitwise check
  ⚙️ Applying pass: AssignToBitwise
    ⚡ Transformed Debug = true → flags |= FlagDebug
  ⚙️ Applying pass: StringObfuscate
    🔒 Obfuscated string literal in APIKey
  ⚙️ Applying pass: JumpTable
    🚀 Converted if-chain to jump table

📋 Context map saved: ./context.map.json
✅ Transpilation complete: 42 files processed
```

---

## **Depois disso**

* Código **compilável** e já otimizado.
* `.map.json` com todos os mapeamentos originais ↔ transpilados.
* Potencial de ganho REAL de performance e footprint.
* Segurança extra contra engenharia reversa.
