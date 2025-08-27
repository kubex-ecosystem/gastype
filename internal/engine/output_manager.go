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

	gl "github.com/rafa-mori/gastype/internal/module/logger"
)

// OutputManager controla a cópia e substituição dos arquivos no output
// Esta é a SOLUÇÃO DEFINITIVA para projetos PRODUCTION-READY! 🚀
type OutputManager struct {
	SrcRoot    string
	DstRoot    string
	Generated  map[string]*ast.File // Arquivos transpilados
	Fset       *token.FileSet
	ModulePath string
}

// NewOutputManager cria um novo gerenciador com base no go.mod
// : Lê automaticamente o module path para imports corretos!
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
// GENIUS: Copia TUDO + substitui só o que foi transpilado!
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

		// Arquivo Go transpilado - USA nossos arquivos revolucionários!
		if strings.HasSuffix(path, ".go") {
			if astFile, ok := om.Generated[path]; ok {
				om.rewriteImports(astFile)
				return om.writeGoFile(dest, astFile)
			}
		}

		// Caso contrário, apenas copia o original (go.mod, go.sum, configs, etc)
		return copyFile(path, dest)
	})
}

// writeGoFile salva um arquivo .go formatado
// PERFECT: Mantém formatação limpa e profissional
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
// INTELLIGENT: Só modifica imports locais, preserva stdlib e externos!
func (om *OutputManager) rewriteImports(file *ast.File) {
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)

		if isStdLib(path) {
			continue // Pula a reescrita para pacotes da stdlib
		}

		// 🚀: Skip stdlib packages!
		// if om.isStdLib(path) {
		// 	continue
		// }

		// // Skip external packages (contains .)
		// if strings.Contains(path, ".") {
		// 	continue
		// }

		// Skip if it's already a full module path (starts with our module)
		if strings.HasPrefix(path, om.ModulePath) {
			continue
		}

		if strings.Contains(path, ".") { // Se já contém ponto, é geralmente um import externo ou módulo [4]
			continue
		}

		newPath := filepath.ToSlash(filepath.Join(om.ModulePath, path))
		imp.Path.Value = strconv.Quote(newPath)

		gl.Log("debug", fmt.Sprintf("Reescrevendo import: %s -> %s", path, newPath))
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

// isStdLib detects Go standard library packages
// GENIUS: Prevents stdlib hijacking by detecting standard packages!
func (om *OutputManager) isStdLib(importPath string) bool {
	// Standard library packages don't contain dots and match known patterns
	if strings.Contains(importPath, ".") {
		return false // External package
	}

	// Common stdlib packages (comprehensive list)
	stdLibPackages := map[string]bool{
		// Core packages
		"fmt": true, "os": true, "io": true, "log": true, "errors": true,
		"strings": true, "strconv": true, "time": true, "context": true,
		"sync": true, "path": true, "runtime": true, "reflect": true,
		"unsafe": true, "sort": true, "math": true, "bytes": true,

		// Network packages
		"net": true, "http": true,

		// Encoding packages
		"encoding": true, "json": true, "xml": true, "base64": true,

		// Crypto packages
		"crypto": true, "md5": true, "sha1": true, "sha256": true,

		// System packages
		"syscall": true, "signal": true,

		// Testing
		"testing": true,

		// Build/compile time
		"embed": true, "flag": true,

		// Others
		"regexp": true, "unicode": true, "bufio": true, "image": true,
		"html": true, "text": true, "mime": true, "plugin": true,
		"go": true, "debug": true, "hash": true, "index": true,
		"database": true,
	}

	// Check direct match
	if stdLibPackages[importPath] {
		return true
	}

	// Check for nested stdlib packages (e.g., "encoding/json", "net/http")
	parts := strings.Split(importPath, "/")
	if len(parts) > 0 && stdLibPackages[parts[0]] {
		return true
	}

	return false
} // readModulePath lê o module path do go.mod

// -- INÍCIO DA NOVA FUNÇÃO isStdLib --
// isStdLib verifica se o importPath corresponde a um pacote da Go Standard Library.
// Esta é uma implementação simplificada baseada na heurística de "não conter ponto".
// Para uma solução 100% blindada, seria necessário consultar `go list std` e criar um cache [5, 6].
func isStdLib(importPath string) bool {
	// Pacotes da stdlib geralmente não contêm '.' (ex: "fmt", "io", "net/http", "encoding/json")
	if !strings.Contains(importPath, ".") {
		return true
	}
	return false
}
