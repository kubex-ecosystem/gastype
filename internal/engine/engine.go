// Package transpiler provides a modular engine for Go AST transformations
package transpiler

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/rafa-mori/gastype/internal/astutil"

	gl "github.com/rafa-mori/gastype/internal/module/logger"
)

// Engine coordinates passes and context for transpilation
type Engine struct {
	Ctx    *astutil.TranspileContext
	Passes []TranspilePass
}

// TranspilePass interface for any AST transformation
type TranspilePass interface {
	Name() string
	Apply(file *ast.File, fset *token.FileSet, ctx *astutil.TranspileContext) error
}

// DiscoverGoFiles discovers all Go files recursively
func DiscoverGoFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// NewEngine creates a new transpilation engine
func NewEngine(ctx *astutil.TranspileContext) *Engine {
	return &Engine{
		Ctx:    ctx,
		Passes: make([]TranspilePass, 0),
	}
}

// AddPass adds a transpilation pass to the engine
func (e *Engine) AddPass(pass TranspilePass) {
	e.Passes = append(e.Passes, pass)
}

// Run executes the engine on the specified root path
func (e *Engine) Run(root string) error {
	files, err := DiscoverGoFiles(root)
	if err != nil {
		gl.Log("error", fmt.Sprintf("failed to discover Go files: %v", err))
		return fmt.Errorf("failed to discover Go files: %w", err)
	}

	gl.Log("info", fmt.Sprintf("🚀 Starting transpilation engine on %d files\n", len(files)))

	processedFiles := 0
	transformedFiles := 0

	for _, filePath := range files {
		gl.Log("info", fmt.Sprintf("🔍 Processing %s\n", filePath))
		// 🚀: Use shared FileSet from context
		astFile, err := parser.ParseFile(e.Ctx.Fset, filePath, nil, parser.ParseComments)
		if err != nil {
			gl.Log("error", fmt.Sprintf("  ⚠️  Failed to parse %s: %v\n", filePath, err))
			continue
		}

		fileTransformed := false
		for _, pass := range e.Passes {
			gl.Log("info", fmt.Sprintf("  ⚙️  Applying pass: %s\n", pass.Name()))
			// 🚀: Use shared FileSet in passes
			if err := pass.Apply(astFile, e.Ctx.Fset, e.Ctx); err != nil {
				gl.Log("error", fmt.Sprintf("  ⚠️  Pass %s failed on %s: %v\n", pass.Name(), filePath, err))
				return fmt.Errorf("pass %s failed on %s: %w", pass.Name(), filePath, err)
			}
			fileTransformed = true
		}

		if fileTransformed {
			transformedFiles++
			// 🚀: Store transformed files for OutputManager
			e.Ctx.GeneratedFiles[filePath] = astFile
			gl.Log("info", fmt.Sprintf("  ✅ File transformed and stored: %s\n", filePath))
		}

		processedFiles++
	}

	gl.Log("info", fmt.Sprintf("📊 Engine summary: %d files processed, %d transformed\n", processedFiles, transformedFiles))
	gl.Log("info", fmt.Sprintf("🎯 Ready for OutputManager: %d files stored\n", len(e.Ctx.GeneratedFiles)))

	// Save context map if configured
	if e.Ctx.MapFile != "" {
		if err := e.Ctx.SaveMap(); err != nil {
			gl.Log("error", fmt.Sprintf("failed to save context map: %v", err))
			return fmt.Errorf("failed to save context map: %w", err)
		}
		gl.Log("info", fmt.Sprintf("📋 Context map saved: %s\n", e.Ctx.MapFile))
	}

	return nil
}

// GetPassByName returns a pass by its name
func (e *Engine) GetPassByName(name string) TranspilePass {
	for _, pass := range e.Passes {
		if pass.Name() == name {
			return pass
		}
	}
	return nil
}

// ListPasses returns the names of all registered passes
func (e *Engine) ListPasses() []string {
	names := make([]string, len(e.Passes))
	for i, pass := range e.Passes {
		names[i] = pass.Name()
	}
	return names
}
