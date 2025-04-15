package type_check

import (
	"encoding/json"
	"fmt"
	g "github.com/faelmori/gastype/internal/globals"
	t "github.com/faelmori/gastype/types"
	l "github.com/faelmori/logz"
	"os"
	"sync"
	"time"
)

type GasTypeResponse struct {
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
	Info     []string `json:"info"`
}

type TypeCheckSpecInfo struct {
	Definitions     map[string]interface{} `json:"definitions"`
	Functions       map[string]interface{} `json:"functions"`
	Variables       map[string]interface{} `json:"variables"`
	Types           map[string]interface{} `json:"types"`
	Imports         map[string]interface{} `json:"imports"`
	Exports         map[string]interface{} `json:"exports"`
	Constants       map[string]interface{} `json:"constants"`
	Interfaces      map[string]interface{} `json:"interfaces"`
	Structs         map[string]interface{} `json:"structs"`
	Enums           map[string]interface{} `json:"enums"`
	Methods         map[string]interface{} `json:"methods"`
	Routines        map[string]interface{} `json:"routines"`
	GasTypeResponse GasTypeResponse        `json:"gas_type_response"`
}

func NewTypeCheckSpecInfo(info t.IInfo) TypeCheckSpecInfo {
	typeCheckSpecInfo := TypeCheckSpecInfo{
		Definitions:     make(map[string]interface{}),
		Functions:       make(map[string]interface{}),
		Variables:       make(map[string]interface{}),
		Types:           make(map[string]interface{}),
		Imports:         make(map[string]interface{}),
		Exports:         make(map[string]interface{}),
		Constants:       make(map[string]interface{}),
		Interfaces:      make(map[string]interface{}),
		Structs:         make(map[string]interface{}),
		Enums:           make(map[string]interface{}),
		Methods:         make(map[string]interface{}),
		Routines:        make(map[string]interface{}),
		GasTypeResponse: GasTypeResponse{},
	}

	if info == nil {
		return typeCheckSpecInfo
	}

	//typeCheckSpecInfo.Functions = info.GetFunctions()
	//typeCheckSpecInfo.GasTypeResponse = info.GetGasTypeResponse()
	//typeCheckSpecInfo.Variables = info.GetVariables()
	//typeCheckSpecInfo.Types = info.GetTypes()
	//typeCheckSpecInfo.Imports = info.GetImports()
	//typeCheckSpecInfo.Exports = info.GetExports()
	//typeCheckSpecInfo.Constants = info.GetConstants()
	//typeCheckSpecInfo.Interfaces = info.GetInterfaces()
	//typeCheckSpecInfo.Structs = info.GetStructs()
	//typeCheckSpecInfo.Enums = info.GetEnums()
	//typeCheckSpecInfo.Methods = info.GetMethods()
	//typeCheckSpecInfo.Routines = info.GetRoutines()
	//typeCheckSpecInfo.GasTypeResponse = info.GetGasTypeResponse()

	return typeCheckSpecInfo
}

type TypeCheckDetails struct {
	ID              string            `json:"id"`
	Index           int               `json:"index"`
	Context         string            `json:"context"`
	Package         string            `json:"package"`
	Status          string            `json:"status"`
	Lines           int               `json:"lines"`
	StatusCode      int               `json:"status_code"`
	StatusText      string            `json:"status_text"`
	ASTFile         string            `json:"ast_file"`
	AST             interface{}       `json:"ast"`
	Info            TypeCheckSpecInfo `json:"info"`
	GasTypeResponse GasTypeResponse   `json:"gas_type_response"`
}

func NewTypeCheckDetails() TypeCheckDetails {
	return TypeCheckDetails{
		ID:         "",
		Package:    "",
		Status:     "",
		Lines:      0,
		StatusCode: 0,
		StatusText: "",
		ASTFile:    "",
		AST:        nil,
	}
}

type TypeCheckSummary struct {
	TotalPackages  int    `json:"total_packages"`
	Successful     int    `json:"successful"`
	Errors         int    `json:"errors"`
	CriticalErrors int    `json:"critical_errors"` // Nova métrica
	LinesAnalyzed  int    `json:"lines_analyzed"`
	AvgLinesPerPkg int    `json:"avg_lines_per_package"` // Média por pacote
	Resume         string `json:"resume"`
	Score          int    `json:"score"`
}

type TypeCheckResult struct {
	mu         sync.RWMutex
	logger     l.Logger
	chanResult chan t.IResult
	chanError  chan error
	chanDone   chan bool
	worker     t.IWorker
	config     t.IConfig
	Summary    TypeCheckSummary   `json:"summary"`
	Details    []TypeCheckDetails `json:"details"`
	Timestamp  string             `json:"timestamp"`
	Duration   string             `json:"duration"`
}

func NewTypeCheckResult(processConfig g.ProcessConfig) *TypeCheckResult {
	if processConfig.Logger == nil {
		processConfig.Logger = l.GetLogger("GasType")
	}
	if processConfig.ChanResult == nil {
		processConfig.ChanResult = make(chan t.IResult, 10)
	}
	if processConfig.ChanError == nil {
		processConfig.ChanError = make(chan error, 10)
	}
	if processConfig.ChanDone == nil {
		processConfig.ChanDone = make(chan bool, 2)
	}
	typeCheckResult := TypeCheckResult{
		mu:         sync.RWMutex{},
		logger:     processConfig.Logger,
		chanResult: processConfig.ChanResult,
		chanError:  processConfig.ChanError,
		chanDone:   processConfig.ChanDone,
		worker:     processConfig.Worker,
		config:     processConfig.Config,
		Summary: TypeCheckSummary{
			TotalPackages: 0,
			Successful:    0,
			Errors:        0,
			LinesAnalyzed: 0,
			Resume:        "",
			Score:         100, // Por padrão, começa com "100" e ajusta ao longo da execução
		},
		Details:   []TypeCheckDetails{},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	typeCheckResult.Summary.TotalPackages = len(processConfig.Packages)
	typeCheckResult.Summary.LinesAnalyzed = 0
	typeCheckResult.Summary.Errors = 0
	typeCheckResult.Summary.Successful = 0
	typeCheckResult.Summary.Score = 100
	typeCheckResult.Summary.Resume = ""
	typeCheckResult.Timestamp = time.Now().Format(time.RFC3339)
	typeCheckResult.Duration = "0s"
	typeCheckResult.Details = make([]TypeCheckDetails, 0)

	return &typeCheckResult
}

func (tc *TypeCheckResult) AnalyzePackage(pkgName string, lines int, astFile string) TypeCheckDetails {
	// Simulação de análise
	errors := make([]string, 0)
	warnings := make([]string, 0)
	info := []string{"Analysis completed"}

	if lines < 10 { // Exemplo de condição que gera erro
		errors = append(errors, "Package too small")
	}

	return TypeCheckDetails{
		ID:         pkgName,
		Package:    pkgName,
		Status:     "Success",
		Lines:      lines,
		StatusCode: 200,
		StatusText: "OK",
		ASTFile:    astFile,
		Info: TypeCheckSpecInfo{
			Functions: map[string]interface{}{"exampleFunc": "func()"},
			GasTypeResponse: GasTypeResponse{
				Errors:   errors,
				Warnings: warnings,
				Info:     info,
			},
		},
	}
}

func (tc *TypeCheckResult) UpdateSummary() {
	tc.Summary.TotalPackages = len(tc.Details)
	tc.Summary.Errors = 0
	tc.Summary.Successful = 0
	tc.Summary.CriticalErrors = 0
	tc.Summary.LinesAnalyzed = 0
	tc.Summary.AvgLinesPerPkg = 0
	tc.Summary.Score = 1000

	for _, detail := range tc.Details {
		tc.Summary.LinesAnalyzed += detail.Lines
		if detail.Status == "Success" {
			tc.Summary.Successful++
		} else if detail.Status == "Error" {
			tc.Summary.Errors++
			tc.Summary.Score -= 150
		} else if detail.Status == "CriticalError" {
			tc.Summary.CriticalErrors++
			tc.Summary.Score -= 250
		} else if detail.Status == "Warning" {
			tc.Summary.Errors++
			tc.Summary.Score -= 50
		} else if detail.Status == "Info" {
			tc.Summary.Successful++
		} else if detail.Status == "Unknown" {
			tc.Summary.Errors++
			tc.Summary.Score -= 50
		} else {
			tc.Summary.Errors++
			tc.Summary.Score -= 100
		}
		if tc.Summary.Score < 0 {
			tc.Summary.Score = 0
		}
	}

	tc.Summary.AvgLinesPerPkg = tc.Summary.LinesAnalyzed / tc.Summary.TotalPackages

	tc.Summary.Resume = fmt.Sprintf(
		"Packages: %d | Success: %d | Errors: %d | Lines: %d | Score: %d",
		tc.Summary.TotalPackages, tc.Summary.Successful,
		tc.Summary.Errors, tc.Summary.LinesAnalyzed,
		tc.Summary.Score,
	)
}

func (tc *TypeCheckResult) GenerateFinalResult() TypeCheckResult {
	for _, detail := range tc.Details {
		tc.Details = append(tc.Details, detail)
	}

	tc.UpdateSummary()

	return *tc
}

func (tc *TypeCheckResult) SaveResultsToJSON(result TypeCheckResult, outputFile string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outputFile, data, 0644)
}

func (tc *TypeCheckResult) ListenResults() {
	for {
		select {
		case result := <-tc.chanResult:
			tc.logger.InfoCtx(fmt.Sprintf("Processing result: %v", result), nil)
			tc.mu.Lock()
			defer tc.mu.Unlock()
			newDetail := NewTypeCheckDetails()
			newDetail.ID = result.GetPackage()
			newDetail.Package = result.GetPackage()
			newDetail.Status = result.GetStatus()
			newDetail.StatusCode = result.GetStatusCode()
			newDetail.StatusText = result.GetStatusText()
			newDetail.ASTFile = result.GetAstFile()
			newDetail.AST = result.GetAst()
			typeCheckSpecInfo := NewTypeCheckSpecInfo(result.GetInfo())

			newDetail.Info = typeCheckSpecInfo
			tc.Details = append(tc.Details, newDetail)
			tc.UpdateSummary()
		case err := <-tc.chanError:
			tc.logger.ErrorCtx(fmt.Sprintf("Error received: %v", err), nil)
			tc.mu.Lock()
			defer tc.mu.Unlock()
			newDetail := NewTypeCheckDetails()
			newDetail.ID = "Error"
			newDetail.Package = "Error"
			newDetail.Status = "Error"
			newDetail.StatusCode = 500
			newDetail.StatusText = "Internal Server Error"
			newDetail.ASTFile = "Error"
			newDetail.AST = nil
			newDetail.Info = TypeCheckSpecInfo{
				Functions: map[string]interface{}{"error": err.Error()},
				GasTypeResponse: GasTypeResponse{
					Errors:   []string{err.Error()},
					Warnings: []string{},
					Info:     []string{"Error occurred"},
				},
			}
			tc.Details = append(tc.Details, newDetail)
			tc.UpdateSummary()
		case <-tc.chanDone:
			tc.logger.InfoCtx("Processing done signal received", nil)
			tc.UpdateSummary()
			tc.logger.InfoCtx(fmt.Sprintf("Final summary: %v", tc.Summary), nil)
			tc.logger.InfoCtx(fmt.Sprintf("Final details: %v", tc.Details), nil)
			if saveErr := tc.SaveResultsToJSON(*tc, "final_result.json"); saveErr != nil {
				tc.logger.ErrorCtx(fmt.Sprintf("Error saving results: %v", saveErr), nil)
			} else {
				tc.logger.InfoCtx("Results saved to final_result.json", nil)
			}
			return
		}
	}
}

func ProcessPackages(
	processConfig g.ProcessConfig,
) TypeCheckResult {
	typeCheckResult := NewTypeCheckResult(processConfig)

	for _, pkg := range processConfig.Packages {
		astFile := fmt.Sprintf("%s.ast", pkg)
		lines := 0
		detail := typeCheckResult.AnalyzePackage(pkg, lines, astFile)
		typeCheckResult.Details = append(typeCheckResult.Details, detail)
		typeCheckResult.UpdateSummary()
	}

	typeCheckResult.Duration = fmt.Sprintf("%v", time.Since(time.Now()))
	typeCheckResult.Summary.Resume = fmt.Sprintf(
		"Packages: %d | Success: %d | Errors: %d | Lines: %d | Score: %d",
		typeCheckResult.Summary.TotalPackages, typeCheckResult.Summary.Successful,
		typeCheckResult.Summary.Errors, typeCheckResult.Summary.LinesAnalyzed,
		typeCheckResult.Summary.Score,
	)

	// Save results to JSON file
	return typeCheckResult.GenerateFinalResult()
}
