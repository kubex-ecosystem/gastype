package actions

import (
	"fmt"
	g "github.com/faelmori/gastype/internal/globals"
	t "github.com/faelmori/gastype/types"
	"github.com/faelmori/gastype/utils"
	l "github.com/faelmori/logz"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// TypeCheckAction defines a type-checking action
type TypeCheckAction struct {
	Action         // Embedding the base action
	mu             sync.Mutex
	logger         l.Logger
	Config         t.IConfig
	ParsedFiles    map[string][]*ast.File
	files          []string
	Errors         []error
	FileSet        *token.FileSet
	results        map[string]g.Result
	ResultsChannel chan t.IResult
	ErrorChannel   chan error
}

// NewTypeCheckAction creates a new type-checking action
func NewTypeCheckAction(fileList []string, cfg t.IConfig, logger l.Logger) t.IAction {
	fileNameList := make([]string, 0)
	parsedFiles := make(map[string][]*ast.File)
	if len(fileList) > 0 {
		fileNameList = fileList
	} else {
		fileNameList = make([]string, 0)
	}
	if logger == nil {
		logger = l.GetLogger("GasType")
	}
	return &TypeCheckAction{
		Action: Action{
			Type:    "TypeCheck",
			Status:  "Pending",
			Results: make(map[string]t.IResult),
		},
		logger:         logger,
		Config:         cfg,
		files:          fileNameList,
		Errors:         make([]error, 0),
		ParsedFiles:    parsedFiles,
		FileSet:        token.NewFileSet(),
		ErrorChannel:   make(chan error, 50),
		ResultsChannel: make(chan t.IResult, 50),
	}
}

func (tca *TypeCheckAction) GetResults() map[string]t.IResult {
	tca.mu.Lock()
	defer tca.mu.Unlock()
	if tca.Results == nil {
		tca.Results = make(map[string]t.IResult)
	}
	return tca.Results
}

// Execute runs the type-checking process
func (tca *TypeCheckAction) Execute() error {
	tca.logger.NoticeCtx("Starting type checking....", nil)
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
		tca.logger.ErrorCtx(fmt.Sprintf("directory does not exist: %s", absDir), nil)
		return fmt.Errorf("directory does not exist: %s", absDir)
	}

	if tca.ParsedFiles == nil || len(tca.ParsedFiles) == 0 {
		tca.files = make([]string, 0)
		filesErr := utils.CollectGoFiles(absDir, &tca.files, tca.logger)
		if filesErr != nil {
			tca.logger.ErrorCtx(fmt.Sprintf("Error collecting Go files: %s", filesErr.Error()), nil)
			return filesErr
		}
	}

	go tca.listenExecution()

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

			tca.logger.NoticeCtx(fmt.Sprintf("Parsing file: %s", file), nil)
			if err := tca.parseFile(file); err != nil {
				tca.logger.ErrorCtx(fmt.Sprintf("Error parsing file %s: %v", file, err), nil)
				tca.ErrorChannel <- fmt.Errorf("error parsing file %s: %v", file, err)
				return
			}
			tca.logger.NoticeCtx(fmt.Sprintf("Parsed file: %s", file), nil)
		}(file, &wg)
	}
	tca.logger.NoticeCtx("Waiting for all files to be parsed", nil)

	// Wait for all goroutines to finish
	tca.logger.NoticeCtx(fmt.Sprintf("Waiting for all %d routines to finish", rCounter), nil)
	wg.Wait()

	sortedFileList := make([]string, 0)
	if len(tca.files) > 0 {
		for _, file := range tca.files {
			if filepath.Ext(file) == ".go" {
				sortedFileList = append(sortedFileList, file)
			}
		}
		sort.Slice(sortedFileList, func(i, j int) bool {
			return len(tca.ParsedFiles[sortedFileList[i]]) < len(tca.ParsedFiles[sortedFileList[j]])
		})
	} else {
		l.ErrorCtx("No Go files found", nil)
		return nil
	}

	close(tca.ErrorChannel)
	tca.logger.NoticeCtx("All files parsed successfully", nil)

	// Collect errors
	for err := range tca.ErrorChannel {
		l.ErrorCtx(fmt.Sprintf("Error during type checking: %v", err), nil)
		tca.Errors = append(tca.Errors, err)
	}

	tca.logger.NoticeCtx("Type checking completed", nil)
	tca.Status = "Completed"
	return nil
}

// parseFile parses a single Go file
func (tca *TypeCheckAction) parseFile(file string) error {
	tca.mu.Lock()
	defer tca.mu.Unlock()

	tca.logger.NoticeCtx(fmt.Sprintf("Reading file: %s", file), nil)
	src, err := os.ReadFile(file)
	if err != nil {
		l.ErrorCtx(fmt.Sprintf("error reading file %s: %v", file, err), nil)
		return fmt.Errorf("error reading file %s: %v", file, err)
	}

	tca.logger.NoticeCtx(fmt.Sprintf("Parsing file: %s", file), nil)

	if node, nodeErr := parser.ParseFile(tca.FileSet, file, src, parser.AllErrors); nodeErr != nil {
		l.ErrorCtx(fmt.Sprintf("error parsing file %s: %v", file, nodeErr), nil)
		tca.ErrorChannel <- fmt.Errorf("error parsing file %s: %v", file, nodeErr)
	} else {
		if node == nil {
			tca.logger.ErrorCtx(fmt.Sprintf("Parsed node is nil for file %s", file), nil)
			return fmt.Errorf("parsed node is nil for file %s", file)
		} else {
			tca.logger.NoticeCtx(fmt.Sprintf("Parsed file: %s", file), nil)
			if node.Name != nil {
				tca.logger.NoticeCtx(fmt.Sprintf("Parsed node name: %s", node.Name.Name), nil)
			} else {
				tca.logger.ErrorCtx(fmt.Sprintf("Parsed node name is nil for file %s", file), nil)
				return fmt.Errorf("parsed node name is nil for file %s", file)
			}
			if nm, ok := tca.ParsedFiles[node.Name.Name]; !ok {
				tca.ParsedFiles[node.Name.String()] = nm
			} else {
				tca.ParsedFiles[node.Name.String()] = append(tca.ParsedFiles[node.Name.String()], node)
			}
			tca.ResultsChannel <- g.NewResult(node.Name.Name, "Success ✅", nil)
			tca.logger.NoticeCtx(fmt.Sprintf("Parsed file: %s", node.Name.Name), nil)
		}
	}
	return nil
}

func (tca *TypeCheckAction) listenExecution() {
	defer func() {
		tca.logger.NoticeCtx("Execution listener stopped", nil)
		close(tca.ResultsChannel)
	}()
	for {
		select {
		case err := <-tca.ErrorChannel:
			if err != nil {
				tca.logger.ErrorCtx(fmt.Sprintf("Error during type checking: %v", err), nil)
			}
		case result := <-tca.ResultsChannel:
			if result != nil {
				tca.logger.NoticeCtx(fmt.Sprintf("Result: %s", result.GetStatus()), nil)
				if iResult, ok := result.(t.IResult); ok {
					tca.logger.NoticeCtx(fmt.Sprintf("Result: %s", iResult.GetStatus()), nil)
					if iResult.GetError() != "" {
						tca.logger.ErrorCtx(fmt.Sprintf("Error during type checking: %v", iResult.GetError()), nil)
						tca.Errors = append(tca.Errors, fmt.Errorf("error during type checking: %v", iResult.GetError()))
					} else {
						tca.logger.NoticeCtx(fmt.Sprintf("Parsed file: %s", iResult.GetPackage()), nil)
						tca.Results[iResult.GetPackage()] = iResult
					}
				} else {
					tca.logger.ErrorCtx(fmt.Sprintf("Error casting result to IResult: %v", result), nil)
				}
			}
		}
	}
}
