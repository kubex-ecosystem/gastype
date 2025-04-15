package actions

import (
	"fmt"
	g "github.com/faelmori/gastype/internal/globals"
	t "github.com/faelmori/gastype/types"
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
	DoneChannel    chan struct{}
}

// NewTypeCheckAction creates a new type-checking action
func NewTypeCheckAction(file string, cfg t.IConfig, logger l.Logger) t.IAction {
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
		files:          []string{file},
		Errors:         make([]error, 0),
		ParsedFiles:    make(map[string][]*ast.File),
		FileSet:        token.NewFileSet(),
		ErrorChannel:   make(chan error, 50),
		ResultsChannel: make(chan t.IResult, 50),
		DoneChannel:    make(chan struct{}, 2),
	}
}

// GetResults returns the results of the type-checking action
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
	tca.logger.DebugCtx("Starting type checking....", nil)

	tca.Status = "Running"
	tca.isRunning = true

	if tca.DoneChannel == nil {
		tca.DoneChannel = make(chan struct{}, 2)
	}

	defer func() {
		tca.DoneChannel <- struct{}{}
	}()

	go tca.listenExecution()

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
			tca.logger.DebugCtx(fmt.Sprintf("Parsing file: %s", file), nil)
			if err := tca.parseFile(file); err != nil {
				tca.logger.ErrorCtx(fmt.Sprintf("Error parsing file %s: %v", file, err), nil)
				tca.ErrorChannel <- fmt.Errorf("error parsing file %s: %v", file, err)
				return
			}
			tca.logger.DebugCtx(fmt.Sprintf("Parsed file: %s", file), nil)
		}(file, &wg)
	}
	tca.logger.DebugCtx("Waiting for all files to be parsed", nil)

	// Wait for all goroutines to finish
	tca.logger.DebugCtx(fmt.Sprintf("Waiting for all %d routines to finish", rCounter), nil)
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
	return nil
}

// GetResultsChannel returns the results channel
func (tca *TypeCheckAction) GetResultsChannel() chan t.IResult {
	if tca.ResultsChannel == nil {
		tca.ResultsChannel = make(chan t.IResult, 50)
	}
	return tca.ResultsChannel
}

// GetErrorChannel returns the error channel
func (tca *TypeCheckAction) GetErrorChannel() chan error {
	if tca.ErrorChannel == nil {
		tca.ErrorChannel = make(chan error, 50)
	}
	return tca.ErrorChannel
}

// GetDoneChannel returns the done channel
func (tca *TypeCheckAction) GetDoneChannel() chan struct{} {
	if tca.DoneChannel == nil {
		tca.DoneChannel = make(chan struct{}, 2)
	}
	return tca.DoneChannel
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
		tca.ErrorChannel <- fmt.Errorf("error parsing file %s: %v", file, nodeErr)
	} else {
		if node == nil {
			tca.logger.ErrorCtx(fmt.Sprintf("Parsed node is nil for file %s", file), nil)
			return fmt.Errorf("parsed node is nil for file %s", file)
		} else {
			tca.logger.DebugCtx(fmt.Sprintf("Parsed info from file: %s", file), nil)
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
			tca.ResultsChannel <- g.NewResult(node.Name.Name, "Success ✅", nil, node)
			tca.logger.DebugCtx(fmt.Sprintf("Parsed file: %s", node.Name.Name), nil)
		}
	}

	return nil
}

// listenExecution listens for errors and results from the type-checking process
func (tca *TypeCheckAction) listenExecution() {
	defer func() {
		tca.logger.DebugCtx("Execution listener stopped", nil)
		close(tca.ResultsChannel)
	}()
	for {
		select {
		case err := <-tca.ErrorChannel:
			if err != nil {
				tca.logger.ErrorCtx(fmt.Sprintf("Error during type checking: %v", err), nil)
				tca.Errors = append(tca.Errors, err)
				tca.Status = "Error"
				tca.ResultsChannel <- g.NewResult("", "Error ❌", err, nil)
			} else {
				tca.logger.DebugCtx("No error received, continuing...", nil)
			}
		case result := <-tca.ResultsChannel:
			if result != nil {
				if iResult, ok := result.(t.IResult); ok {
					strResult := tca.consolidateFileInfo(iResult)
					tca.logger.InfoCtx(strResult, nil)
					tca.Results[iResult.GetPackage()] = iResult
				} else {
					tca.logger.ErrorCtx(fmt.Sprintf("Error casting result to IResult: %v", result), nil)
				}
			}
		case <-tca.DoneChannel:
			tca.logger.DebugCtx("Action completed", nil)
			tca.Status = "Completed"

			tca.logger.DebugCtx("Closing channels", nil)
			close(tca.DoneChannel)

			tca.isRunning = false

			return
		}
	}
}

func (tca *TypeCheckAction) consolidateFileInfo(iResult t.IResult) string {
	info := iResult.GetInfo()

	var filePath string
	var fileInfo os.FileInfo
	var astFile *ast.File
	var fileVersion, fileName string
	//var fileSize int64

	var aFl *ast.File
	if iResult.GetAst() != nil {
		if astFl, ok := iResult.GetAst().(*ast.File); !ok {
			tca.logger.ErrorCtx(fmt.Sprintf("Error casting AST file: %v", iResult.GetAst()), nil)
			return ""
		} else {
			aFl = astFl
		}
	} else {
		tca.logger.ErrorCtx("AST file is nil", nil)
		return ""
	}
	if aFl != nil {
		astFile = aFl
	}

	fileVersion = info.GetFileVersions()[astFile]
	if fileVersion == "" {
		fileVersion = "unknown"
	}
	if filePath = astFile.Name.Name; filePath == "" {
		var filePathErr error
		filePath, filePathErr = filepath.Abs(filepath.Join(tca.Config.GetDir(), astFile.Name.Name))
		if filePathErr != nil {
			tca.logger.ErrorCtx(fmt.Sprintf("Error getting absolute path: %s", filePathErr), nil)
			filePath = astFile.Name.Name
		}
	} else {
		tca.logger.DebugCtx(fmt.Sprintf("File path: %s", filePath), nil)
		filePath, _ = filepath.Abs(filePath)
	}

	if fInfo, err := os.Stat(fileName); err == nil {
		if fInfo != nil {
			fileInfo = fInfo
		}
		fileName = filepath.Base(filePath)
		filePath = fileInfo.Name()
		fileVersion = fileInfo.ModTime().Format("2006-01-02 15:04:05")
		//fileSize = fileInfo.Size() / 1024
	} /* else if os.IsNotExist(err) {
		//tca.logger.ErrorCtx(fmt.Sprintf("File does not exist: %s", filePath), nil)
	} else {
		//tca.logger.ErrorCtx(fmt.Sprintf("Error getting file info: %v", err), nil)
	}*/

	result := fmt.Sprintf("[Package: %s] Valid %s, Status: %s, Lines: %d", astFile.Name.Name, fileName, iResult.GetStatus(), len(astFile.Decls))
	//result += fmt.Sprintf("[Package: %s] Name: %s, Version %s, Size %d KB", astFile.Name.Name, filePath, fileVersion, fileSize)

	tca.logger.DebugCtx(result, nil)

	return result
}
