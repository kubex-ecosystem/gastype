// Package transpiler provides a modular engine for Go AST transformations
package transpiler

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Engine coordinates passes and context for transpilation
type Engine struct {
	Ctx    *TranspileContext
	Passes []TranspilePass
}

// TranspilePass interface for any AST transformation
type TranspilePass interface {
	Name() string
	Apply(file *ast.File, fset *token.FileSet, ctx *TranspileContext) error
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
func NewEngine(ctx *TranspileContext) *Engine {
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
		return fmt.Errorf("failed to discover Go files: %w", err)
	}

	fmt.Printf("🚀 Starting transpilation engine on %d files\n", len(files))

	processedFiles := 0
	transformedFiles := 0

	for _, filePath := range files {
		fmt.Printf("🔍 Processing %s\n", filePath)
		fset := token.NewFileSet()
		astFile, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			fmt.Printf("  ⚠️  Failed to parse %s: %v\n", filePath, err)
			continue
		}

		fileTransformed := false
		for _, pass := range e.Passes {
			fmt.Printf("  ⚙️  Applying pass: %s\n", pass.Name())
			if err := pass.Apply(astFile, fset, e.Ctx); err != nil {
				return fmt.Errorf("pass %s failed on %s: %w", pass.Name(), filePath, err)
			}
			fileTransformed = true
		}

		if fileTransformed {
			transformedFiles++
		}

		if !e.Ctx.DryRun && fileTransformed {
			// Preserve directory structure in output
			relPath, err := filepath.Rel(root, filePath)
			if err != nil {
				relPath = filepath.Base(filePath)
			}

			outPath := filepath.Join(e.Ctx.OutputDir, relPath)
			if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			outFile, err := os.Create(outPath)
			if err != nil {
				return fmt.Errorf("failed to create output file %s: %w", outPath, err)
			}

			if err := printer.Fprint(outFile, fset, astFile); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write transformed file %s: %w", outPath, err)
			}
			outFile.Close()

			fmt.Printf("  ✅ Saved transformed file: %s\n", outPath)
		}

		processedFiles++
	}

	fmt.Printf("📊 Engine summary: %d files processed, %d transformed\n", processedFiles, transformedFiles)

	// Save context map if configured
	if e.Ctx.MapFile != "" {
		if err := e.Ctx.SaveMap(); err != nil {
			return fmt.Errorf("failed to save context map: %w", err)
		}
		fmt.Printf("📋 Context map saved: %s\n", e.Ctx.MapFile)
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
