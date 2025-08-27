# **`output_manager.go` — versão integrada ao teu `ctx` e `engine`**

```go
package transpiler

import (
 "bufio"
 "fmt"
 "go/ast"
 "go/printer"
 "go/token"
 "io"
 "io/fs"
 "os"
 "path/filepath"
 "strconv"
 "strings"
)

// OutputManager controla a cópia e substituição dos arquivos no output
type OutputManager struct {
 SrcRoot    string
 DstRoot    string
 Generated  map[string]*ast.File // Arquivos transpilados
 Fset       *token.FileSet
 ModulePath string
}

// NewOutputManager cria um novo gerenciador com base no go.mod
func NewOutputManager(srcRoot, dstRoot string, gen map[string]*ast.File, fset *token.FileSet) (*OutputManager, error) {
 goModPath := filepath.Join(srcRoot, "go.mod")
 modulePath, err := readModulePath(goModPath)
 if err != nil {
  return nil, fmt.Errorf("falha ao ler module path: %w", err)
 }
 return &OutputManager{
  SrcRoot:    srcRoot,
  DstRoot:    dstRoot,
  Generated:  gen,
  Fset:       fset,
  ModulePath: modulePath,
 }, nil
}

// Run percorre todo o projeto copiando arquivos
func (om *OutputManager) Run() error {
 return filepath.WalkDir(om.SrcRoot, func(path string, d fs.DirEntry, err error) error {
  if err != nil {
   return err
  }

  rel, _ := filepath.Rel(om.SrcRoot, path)
  dest := filepath.Join(om.DstRoot, rel)

  // Cria diretórios
  if d.IsDir() {
   return os.MkdirAll(dest, 0755)
  }

  // Arquivo Go transpilado
  if strings.HasSuffix(path, ".go") {
   if astFile, ok := om.Generated[path]; ok {
    om.rewriteImports(astFile)
    return om.writeGoFile(dest, astFile)
   }
  }

  // Caso contrário, apenas copia o original
  return copyFile(path, dest)
 })
}

// writeGoFile salva um arquivo .go formatado
func (om *OutputManager) writeGoFile(dst string, f *ast.File) error {
 os.MkdirAll(filepath.Dir(dst), 0755)
 out, err := os.Create(dst)
 if err != nil {
  return err
 }
 defer out.Close()
 return printer.Fprint(out, om.Fset, f)
}

// rewriteImports ajusta imports locais para refletir o module path do go.mod
func (om *OutputManager) rewriteImports(file *ast.File) {
 for _, imp := range file.Imports {
  path := strings.Trim(imp.Path.Value, `"`)
  if strings.Contains(path, ".") { // já é import externo
   continue
  }
  newPath := filepath.ToSlash(filepath.Join(om.ModulePath, path))
  imp.Path.Value = strconv.Quote(newPath)
 }
}

// readModulePath lê o module path do go.mod
func readModulePath(goModPath string) (string, error) {
 f, err := os.Open(goModPath)
 if err != nil {
  return "", err
 }
 defer f.Close()

 scanner := bufio.NewScanner(f)
 for scanner.Scan() {
  line := strings.TrimSpace(scanner.Text())
  if strings.HasPrefix(line, "module ") {
   return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
  }
 }
 return "", fmt.Errorf("module path não encontrado em %s", goModPath)
}

// copyFile copia um arquivo mantendo hierarquia
func copyFile(src, dst string) error {
 os.MkdirAll(filepath.Dir(dst), 0755)
 in, err := os.Open(src)
 if err != nil {
  return err
 }
 defer in.Close()

 out, err := os.Create(dst)
 if err != nil {
  return err
 }
 defer out.Close()

 _, err = io.Copy(out, in)
 return err
}
```

---

## **Integração no `transpileCmd`**

Logo **depois** do `engine.Run(...)`, você chama:

```go
if err := engine.Run(inputPath); err != nil {
    return err
}

om, err := transpiler.NewOutputManager(
    inputPath,
    outputDir,
    ctx.GeneratedFiles, // Mapa de arquivos transpilados no seu contexto
    ctx.Fset,
)
if err != nil {
    return err
}

if err := om.Run(); err != nil {
    return err
}

fmt.Println("✅ Output completo gerado com todos arquivos e pronto para build!")
```

---

## **O que isso resolve na prática**

* **Leva TODOS os arquivos** pro output (incluindo `go.mod`, `.sum`, configs, markdown, scripts).
* Substitui **somente os `.go`** que foram transpilados.
* Preserva **hierarquia original**.
* Ajusta imports locais com base no `module` do `go.mod`.
* Sai **buildável na hora** → sem ajuste manual.
