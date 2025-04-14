package actions

import (
	"fmt"
	"github.com/faelmori/gastype/internal/globals"
	"github.com/faelmori/gastype/utils"
	"go/parser"
	"os"
	"path/filepath"
	"sync"

	t "github.com/faelmori/gastype/types"
	l "github.com/faelmori/logz"
	"go/ast"
	"go/token"
	"go/types"
)

type CheckResult struct {
	Package string `json:"package"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

// TypeCheckAction defines a type-checking action
type TypeCheckAction struct {
	Action       // Embedding the base action
	mu           sync.Mutex
	logger       l.Logger
	Config       t.IConfig
	ParsedFiles  map[string][]*ast.File
	files        []string
	Errors       []error
	FileSet      *token.FileSet
	results      chan CheckResult
	ErrorChannel chan error
}

// NewTypeCheckAction creates a new type-checking action
func NewTypeCheckAction(pkg string, files []*ast.File, cfg t.IConfig, logger l.Logger) *TypeCheckAction {
	pkgName := pkg
	fileNameList := make([]string, 0)
	parsedFiles := make(map[string][]*ast.File)
	for _, file := range files {
		if file.Name != nil {
			pkgName = file.Name.Name
			fileNameList = append(fileNameList, file.Name.Name)
		}
		parsedFiles[pkgName] = append(parsedFiles[pkgName], file)
	}
	return &TypeCheckAction{
		Action: Action{
			Type:    "TypeCheck",
			Status:  "Pending",
			Results: make(map[string]interface{}),
		},
		logger:       logger,
		Config:       cfg,
		files:        fileNameList,
		Errors:       make([]error, 0),
		ParsedFiles:  parsedFiles,
		FileSet:      token.NewFileSet(),
		ErrorChannel: make(chan error, cfg.GetWorkerCount()),
	}
}

// Execute runs the type-checking process
func (tca *TypeCheckAction) Execute() error {
	tca.logger.DebugCtx("Starting type checking....", nil)
	tca.Status = "Running"
	defer func() { tca.Status = "Completed" }()

	dir := tca.Config.GetDir()
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("invalid directory path: %v", err)
	}

	// Ensure the directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		tca.logger.ErrorCtx(fmt.Sprintf("directory does not exist: %s", absDir), nil)
		return fmt.Errorf("directory does not exist: %s", absDir)
	}

	parsedFiles := make([]*ast.File, 0)
	if tca.ParsedFiles == nil || len(tca.ParsedFiles) == 0 {
		tca.files = make([]string, 0)
		filesErr := utils.CollectGoFiles(absDir, &tca.files, &parsedFiles, tca.logger)
		if filesErr != nil {
			tca.logger.ErrorCtx(fmt.Sprintf("Error collecting Go files: %s", filesErr.Error()), nil)
			return filesErr
		}
	}

	tca.logger.InfoCtx(fmt.Sprintf("Collected %d Go files", len(tca.files)), nil)
	// Parse files in parallel
	var wg sync.WaitGroup
	var rCounter = 0
	for _, file := range tca.files {
		wg.Add(1)
		rCounter++
		go func(file string, wg *sync.WaitGroup) {
			defer func(wg *sync.WaitGroup) {
				wg.Done()
				rCounter--
			}(wg)
			conf := types.Config{
				Error: func(err error) {
					tca.results <- CheckResult{Package: file, Status: "Error ❌", Error: err.Error()}
				},
			}
			tca.logger.DebugCtx(fmt.Sprintf("Parsing file: %s", file), nil)
			if parseErr := tca.parseFile(file); parseErr != nil {
				tca.logger.ErrorCtx(fmt.Sprintf("Error parsing file %s: %v", file, parseErr), nil)
				tca.ErrorChannel <- parseErr
			}
			tca.logger.DebugCtx(fmt.Sprintf("Finished parsing file: %s", file), nil)
			info := globals.NewInfo()
			if _, err := conf.Check(file, tca.FileSet, parsedFiles, info); err != nil {
				tca.results <- CheckResult{Package: file, Status: "Failed ❌", Error: err.Error()}
			} else {
				tca.results <- CheckResult{Package: file, Status: "Success ✅"}
			}
		}(file, &wg)
	}
	tca.logger.DebugCtx("Waiting for all files to be parsed", nil)

	// Wait for all goroutines to finish
	tca.logger.DebugCtx(fmt.Sprintf("Waiting for all %d routines to finish", rCounter), nil)
	wg.Wait()

	close(tca.ErrorChannel)
	tca.logger.DebugCtx("All files parsed successfully", nil)

	// Collect errors
	for err := range tca.ErrorChannel {
		l.ErrorCtx(fmt.Sprintf("Error during type checking: %v", err), nil)
		tca.Errors = append(tca.Errors, err)
	}

	tca.logger.DebugCtx("Type checking completed", nil)
	tca.Status = "Completed"
	return nil
}

// parseFile parses a single Go file
func (tca *TypeCheckAction) parseFile(file string) error {
	tca.mu.Lock()
	defer tca.mu.Unlock()

	tca.logger.DebugCtx(fmt.Sprintf("Reading file: %s", file), nil)
	src, err := os.ReadFile(file)
	if err != nil {
		l.ErrorCtx(fmt.Sprintf("error reading file %s: %v", file, err), nil)
		return fmt.Errorf("error reading file %s: %v", file, err)
	}

	tca.logger.DebugCtx(fmt.Sprintf("Parsing file: %s", file), nil)

	if node, nodeErr := parser.ParseFile(tca.FileSet, file, src, parser.AllErrors); nodeErr != nil {
		l.ErrorCtx(fmt.Sprintf("error parsing file %s: %v", file, nodeErr), nil)
		tca.Results[node.Name.Name] = globals.NewResult(
			node.Name.Name,
			fmt.Sprintf("error parsing file %s: %v", file, nodeErr),
			nodeErr,
		)
	} else {
		if node == nil {
			tca.logger.ErrorCtx(fmt.Sprintf("Parsed node is nil for file %s", file), nil)
			return fmt.Errorf("parsed node is nil for file %s", file)
		} else {
			tca.logger.DebugCtx(fmt.Sprintf("Parsed file: %s", file), nil)
			if node.Name != nil {
				tca.logger.DebugCtx(fmt.Sprintf("Parsed node name: %s", node.Name.Name), nil)
			} else {
				tca.logger.ErrorCtx(fmt.Sprintf("Parsed node name is nil for file %s", file), nil)
				return fmt.Errorf("parsed node name is nil for file %s", file)
			}
			if nm, ok := tca.ParsedFiles[node.Name.Name]; !ok {
				tca.ParsedFiles[node.Name.String()] = nm
			} else {
				tca.ParsedFiles[node.Name.String()] = append(tca.ParsedFiles[node.Name.String()], node)
			}
		}
	}
	return nil
}

// Error sends an error message to the results channel
func (tca *TypeCheckAction) Error(pkgName string, err error) {
	tca.results <- CheckResult{Package: pkgName, Status: "Error ❌", Error: err.Error()}
}
