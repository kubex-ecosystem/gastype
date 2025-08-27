// Package actions implements various actions for type checking in Go projects.
package actions

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go/ast"
	"go/parser"
	"go/token"

	types "github.com/rafa-mori/gastype/interfaces"
	gl "github.com/rafa-mori/gastype/internal/module/logger"
)

// TypeCheckAction defines a type-checking action
type TypeCheckAction struct {
	Action       // Embedding the base action
	Config       types.IConfig
	ParsedFiles  map[string][]*ast.File
	FileSet      *token.FileSet
	ErrorChannel chan error
}

// NewTypeCheckAction creates a new type-checking action
func NewTypeCheckAction(pkg string, files []*ast.File, cfg types.IConfig) *TypeCheckAction {
	return &TypeCheckAction{
		Action: Action{
			Type:    "TypeCheck",
			Status:  "Pending",
			Results: make(map[string]interface{}),
		},
		Config:       cfg,
		ParsedFiles:  map[string][]*ast.File{pkg: files},
		FileSet:      token.NewFileSet(),
		ErrorChannel: make(chan error, cfg.GetWorkerCount()),
	}
}

// Execute runs the type-checking process
func (tca *TypeCheckAction) Execute() error {
	tca.Status = "Running"
	defer func() { tca.Status = "Completed" }()

	dir := tca.Config.GetDir()
	absDir, err := filepath.Abs(dir)
	if err != nil {
		gl.Log("error", fmt.Sprintf("invalid directory path: %v", err))
		return fmt.Errorf("invalid directory path: %v", err)
	}

	// Ensure the directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		gl.Log("error", fmt.Sprintf("directory does not exist: %s", absDir))
		return fmt.Errorf("directory does not exist: %s", absDir)
	}

	// Find all Go files
	files, err := filepath.Glob(filepath.Join(absDir, "*.go"))
	if err != nil {
		gl.Log("error", fmt.Sprintf("error finding Go files: %v", err))
		return fmt.Errorf("error finding Go files: %v", err)
	}
	if len(files) == 0 {
		gl.Log("error", fmt.Sprintf("no Go files found in directory: %s", absDir))
		return fmt.Errorf("no Go files found in directory: %s", absDir)
	}

	// Parse files in parallel
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			if parseErr := tca.parseFile(file); parseErr != nil {
				tca.ErrorChannel <- parseErr
			}
		}(file)
	}
	wg.Wait()
	close(tca.ErrorChannel)

	// Collect errors
	for err := range tca.ErrorChannel {
		gl.Log("error", fmt.Sprintf("Error during type checking: %v", err))
		tca.Errors = append(tca.Errors, err)
	}

	tca.Status = "Completed"
	return nil
}

// parseFile parses a single Go file
func (tca *TypeCheckAction) parseFile(file string) error {
	src, err := os.ReadFile(file)
	if err != nil {
		gl.Log("error", fmt.Sprintf("error reading file %s: %v", file, err))
		return fmt.Errorf("error reading file %s: %v", file, err)
	}

	node, err := parser.ParseFile(tca.FileSet, file, src, parser.AllErrors)
	if err != nil {
		gl.Log("error", fmt.Sprintf("error parsing %s: %v", file, err))
		return fmt.Errorf("error parsing %s: %v", file, err)
	}

	// Store parsed file
	tca.ParsedFiles[node.Name.Name] = append(tca.ParsedFiles[node.Name.Name], node)
	return nil
}
