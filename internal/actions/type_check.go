package actions

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/faelmori/gastype/types"
	l "github.com/faelmori/logz"
	"go/ast"
	"go/parser"
	"go/token"
)

// TypeCheckAction defines a type-checking action
type TypeCheckAction struct {
	Action       // Embedding the base action
	mu           sync.Mutex
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

func collectGoFiles(dirPath string, files *[]string, lgr l.Logger) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		lgr.Error(fmt.Sprintf("error reading directory %s: %v", dirPath, err), nil)
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			// Recursivamente percorre subpastas
			err := collectGoFiles(fullPath, files, lgr)
			if err != nil {
				lgr.Error(fmt.Sprintf("error reading subdirectory %s: %v", fullPath, err), nil)
				return err
			}
		} else if filepath.Ext(entry.Name()) == ".go" {
			lgr.Info(fmt.Sprintf("Found Go file: %s", fullPath), nil)
			*files = append(*files, fullPath)
		}
	}

	return nil
}

// Execute runs the type-checking process
func (tca *TypeCheckAction) Execute() error {
	l.GetLogger("GasType").Info("Starting type checking", nil)
	// Set the status to running
	// and defer setting it to completed
	// to ensure it is set even if an error occurs
	// and to ensure the status is set to completed
	tca.Status = "Running"
	defer func() { tca.Status = "Completed" }()

	dir := tca.Config.GetDir()
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("invalid directory path: %v", err)
	}

	// Ensure the directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		l.GetLogger("GasType").Error(fmt.Sprintf("directory does not exist: %s", absDir), nil)
		return fmt.Errorf("directory does not exist: %s", absDir)
	}

	files := make([]string, 0)
	filesErr := collectGoFiles(absDir, &files, l.GetLogger("GasType"))
	if filesErr != nil {
		fmt.Println("Erro ao coletar arquivos:", filesErr)
		return filesErr
	}

	// Parse files in parallel
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			l.GetLogger("GasType").Info(fmt.Sprintf("Parsing file: %s", file), nil)
			if parseErr := tca.parseFile(file); parseErr != nil {
				l.GetLogger("GasType").Error(fmt.Sprintf("Error parsing file %s: %v", file, parseErr), nil)
				tca.ErrorChannel <- parseErr
			}
			l.GetLogger("GasType").Info(fmt.Sprintf("Finished parsing file: %s", file), nil)
		}(file)
	}
	l.GetLogger("GasType").Info("Waiting for all files to be parsed", nil)
	wg.Wait()
	close(tca.ErrorChannel)
	l.Info("All files parsed successfully", nil)

	// Collect errors
	for err := range tca.ErrorChannel {
		l.Error(fmt.Sprintf("Error during type checking: %v", err), nil)
		tca.Errors = append(tca.Errors, err)
	}

	l.GetLogger("GasType").Info("Type checking completed", nil)
	tca.Status = "Completed"
	return nil
}

// parseFile parses a single Go file
func (tca *TypeCheckAction) parseFile(file string) error {
	l.GetLogger("GasType").Info(fmt.Sprintf("Reading file: %s", file), nil)
	src, err := os.ReadFile(file)
	if err != nil {
		l.Error(fmt.Sprintf("error reading file %s: %v", file, err), nil)
		return fmt.Errorf("error reading file %s: %v", file, err)
	}

	l.GetLogger("GasType").Info(fmt.Sprintf("Parsing file: %s", file), nil)
	node, err := parser.ParseFile(tca.FileSet, file, src, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("error parsing %s: %v", file, err)
	}

	l.GetLogger("GasType").Info(fmt.Sprintf("Parsed file: %s", file), nil)
	tca.ParsedFiles[node.Name.Name] = append(tca.ParsedFiles[node.Name.Name], node)
	return nil
}
