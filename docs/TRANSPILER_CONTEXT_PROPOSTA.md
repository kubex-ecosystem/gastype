# **Estrutura base**

```go
package transpiler

type TranspileContext struct {
    // Configurações gerais
    Ofuscate  bool   // Se true, nomes e estrutura serão embaralhados
    MapFile   string // Caminho do arquivo .map.json de saída
    InputFile string
    OutputDir string

    // Análise
    Structs map[string]*StructInfo // Struct original → info detalhada
    Flags   map[string][]string    // Struct → lista de flags geradas
}

// Informação sobre cada struct detectada
type StructInfo struct {
    OriginalName string            // Nome original (ex: Config)
    NewName      string            // Nome final (ex: ConfigFlags)
    BoolFields   []string          // Campos bool originais
    FlagMapping  map[string]string // BoolField → FlagName
}
```

---

## **Inicialização**

```go
func NewContext(inputFile, outputDir string, ofuscate bool, mapFile string) *TranspileContext {
    return &TranspileContext{
        Ofuscate:   ofuscate,
        MapFile:    mapFile,
        InputFile:  inputFile,
        OutputDir:  outputDir,
        Structs:    make(map[string]*StructInfo),
        Flags:      make(map[string][]string),
    }
}
```

---

## **Registro de structs**

Durante a análise AST:

```go
func (ctx *TranspileContext) AddStruct(originalName, newName string, boolFields []string) {
    mapping := make(map[string]string)
    for _, f := range boolFields {
        mapping[f] = "Flag" + strings.Title(f)
    }

    ctx.Structs[originalName] = &StructInfo{
        OriginalName: originalName,
        NewName:      newName,
        BoolFields:   boolFields,
        FlagMapping:  mapping,
    }
    ctx.Flags[newName] = boolFields
}
```

---

## **Geração do arquivo `.map.json`**

```go
import (
    "encoding/json"
    "os"
)

func (ctx *TranspileContext) SaveMap() error {
    if ctx.MapFile == "" {
        return nil
    }
    data, err := json.MarshalIndent(ctx, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(ctx.MapFile, data, 0644)
}
```

---

## **Exemplo de saída `.map.json`**

```json
{
  "Ofuscate": false,
  "MapFile": "out/context.map.json",
  "InputFile": "service.go",
  "OutputDir": "out_optimized",
  "Structs": {
    "Config": {
      "OriginalName": "Config",
      "NewName": "ConfigFlags",
      "BoolFields": ["Debug", "Auth"],
      "FlagMapping": {
        "Debug": "FlagDebug",
        "Auth": "FlagAuth"
      }
    }
  },
  "Flags": {
    "ConfigFlags": ["Debug", "Auth"]
  }
}
```

---

## **Integração no CLI**

```go
var mapFile string

func init() {
    transpileCmd.Flags().StringVar(&mapFile, "map", "", "Gera arquivo de mapeamento JSON")
}

var transpileCmd = &cobra.Command{
    Use:   "transpile [arquivo.go]",
    Short: "Transpila bools para bitflags",
    RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) == 0 {
            return fmt.Errorf("forneça um arquivo Go para transpilar")
        }

        ctx := NewContext(args[0], "./out_optimized", false, mapFile)
        err := TranspileBoolToFlagsWithContext(ctx)
        if err != nil {
            return err
        }
        return ctx.SaveMap()
    },
}
```

---

## **Benefícios**

* **Documenta a transpilação** → qualquer dev pode ver original vs convertido.
* **Suporte futuro a rollback** → só aplica se `Ofuscate=false`.
* **Base pra plugins IDE** → highlight de origem original.
* **Debug seguro** → mesmo com AST avançado, você sabe de onde veio cada modificação.
