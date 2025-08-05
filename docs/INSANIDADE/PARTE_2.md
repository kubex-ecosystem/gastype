# **🔥 GasType Engine — Pipeline Fusão Total**

### Passes padrão integrados

1. **BoolToFlagsPass** → detecta structs com bools e cria flags.
2. **IfToBitwisePass** → transforma `if cfg.Bool` em bitwise check.
3. **AssignToBitwisePass** → transforma `cfg.Bool = true/false` em set/clear bitmask.

---

## **`pipeline.go`**

```go
package transpiler

import "strings"

// Cria pipeline padrão com todos os passes core
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
        }
    }
    return selected
}
```

---

## **`cmd/transpile.go`**

```go
var transpileCmd = &cobra.Command{
    Use:   "transpile [pasta|arquivo]",
    Short: "Transpila projeto Go para otimizações bitwise",
    RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) == 0 {
            return fmt.Errorf("forneça um diretório ou arquivo Go")
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

## **Comando mágico único**

```bash
# Rodar tudo de uma vez — bool2flags + if2bitwise + assign2bitwise
gastype transpile ./ \
  --passes bool2flags,if2bitwise,assign2bitwise \
  --map ./context.map.json

# Só simular (sem salvar arquivos)
gastype transpile ./ \
  --passes bool2flags,if2bitwise,assign2bitwise \
  --dry-run

# Estimar ganhos antes de aplicar
gastype transpile ./ \
  --passes bool2flags,if2bitwise,assign2bitwise \
  --estimate-perf --dry-run
```

---

## **Como vai funcionar**

1. **Descobre todos os `.go`** no projeto (`filepath.WalkDir`).
2. Aplica **passo 1** (bool→flags) → cria flags + substitui struct.
3. Aplica **passo 2** (if→bitwise) → troca todos os `if cfg.Bool`.
4. Aplica **passo 3** (assign→bitwise) → troca todos os `cfg.Bool = X`.
5. Gera **código compilável** e `.map.json` consolidado.
