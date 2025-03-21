package actions

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/faelmori/gastype/types"
	log "github.com/faelmori/logz"
	"go/ast"
	"go/parser"
	"go/token"
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
		return fmt.Errorf("invalid directory path: %v", err)
	}

	// Ensure the directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", absDir)
	}

	// Find all Go files
	files, err := filepath.Glob(filepath.Join(absDir, "*.go"))
	if err != nil {
		return fmt.Errorf("error finding Go files: %v", err)
	}
	if len(files) == 0 {
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
		log.Error(fmt.Sprintf("Error during type checking: %v", err), nil)
		tca.Errors = append(tca.Errors, err)
	}

	tca.Status = "Completed"
	return nil
}

// parseFile parses a single Go file
func (tca *TypeCheckAction) parseFile(file string) error {
	src, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", file, err)
	}

	node, err := parser.ParseFile(tca.FileSet, file, src, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("error parsing %s: %v", file, err)
	}

	// Store parsed file
	tca.ParsedFiles[node.Name.Name] = append(tca.ParsedFiles[node.Name.Name], node)
	return nil
}
