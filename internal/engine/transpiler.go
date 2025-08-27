// Package transpiler implements bitwise transpilation pipeline and suggestions
package transpiler

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	gl "github.com/rafa-mori/gastype/internal/module/logger"
)

// StructBitwiseSuggestion represents a suggestion for converting bool fields to flags
type StructBitwiseSuggestion struct {
	StructName string
	BoolFields []string
	File       string
	Line       int
}

// BitwiseTranspiler is the main structure for bitwise transpilation
type BitwiseTranspiler struct {
	fset *token.FileSet
}

// NewBitwiseTranspiler creates a new bitwise transpiler
func NewBitwiseTranspiler() *BitwiseTranspiler {
	return &BitwiseTranspiler{
		fset: token.NewFileSet(),
	}
}

// TranspilationResult represents the result of an analysis/transpilation
type TranspilationResult struct {
	OriginalFile     string            `json:"original_file"`
	TranspiledFile   string            `json:"transpiled_file"`
	Optimizations    []Optimization    `json:"optimizations"`
	SecurityFeatures []SecurityFeature `json:"security_features"`
}

// Optimization represents a found optimization
type Optimization struct {
	Type          string  `json:"type"`
	Description   string  `json:"description"`
	Location      string  `json:"location"`
	BytesSaved    int     `json:"bytes_saved"`
	SpeedupFactor float64 `json:"speedup_factor"`
}

// SecurityFeature represents an applied security feature
type SecurityFeature struct {
	Description string `json:"description"`
	Strength    string `json:"strength"`
}

// AnalyzeFile analyzes a single file and returns optimization results
func (bt *BitwiseTranspiler) AnalyzeFile(filename string) (*TranspilationResult, error) {
	suggestions, err := bt.analyzeFileSugereBitwise(filename)
	if err != nil {
		gl.Log("error", fmt.Sprintf("erro na análise bitwise: %v", err))
		return nil, fmt.Errorf("erro na análise bitwise: %w", err)
	}

	result := &TranspilationResult{
		OriginalFile:     filename,
		TranspiledFile:   strings.Replace(filename, ".go", "_bitwise.go", 1),
		Optimizations:    []Optimization{},
		SecurityFeatures: []SecurityFeature{},
	}

	// Converter sugestões em otimizações
	for _, s := range suggestions {
		opt := Optimization{
			Type:          "BitfieldConversion",
			Description:   fmt.Sprintf("Convert %d bool fields to bitwise flags in struct %s", len(s.BoolFields), s.StructName),
			Location:      fmt.Sprintf("%s:%d", s.File, s.Line),
			BytesSaved:    len(s.BoolFields) * 7, // 8 bytes -> 1 byte per bool
			SpeedupFactor: 2.5,                   // Estimativa conservadora
		}
		result.Optimizations = append(result.Optimizations, opt)

		// Adicionar recurso de segurança
		security := SecurityFeature{
			Description: fmt.Sprintf("Obfuscated struct %s with bitwise operations", s.StructName),
			Strength:    "Medium",
		}
		result.SecurityFeatures = append(result.SecurityFeatures, security)
	}

	return result, nil
}

// AnalyzeProject analyzes an entire project (directory)
func (bt *BitwiseTranspiler) AnalyzeProject(projectPath string) error {
	var results []TranspilationResult

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Analyze only .go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		result, err := bt.AnalyzeFile(path)
		if err != nil {
			gl.Log("error", fmt.Sprintf("error analyzing %s: %v", path, err))
			return fmt.Errorf("error analyzing %s: %w", path, err)
		}

		// Only add if there are optimizations
		if len(result.Optimizations) > 0 {
			results = append(results, *result)
		}

		return nil
	})

	if err != nil {
		gl.Log("error", fmt.Sprintf("error traversing project: %v", err))
		return fmt.Errorf("error traversing project: %w", err)
	}

	return nil
}

// analyzeFileSugereBitwise analyzes a Go file and returns structs that can be converted to bitwise
func (bt *BitwiseTranspiler) analyzeFileSugereBitwise(filename string) ([]StructBitwiseSuggestion, error) {
	file, err := os.Open(filename)
	if err != nil {
		gl.Log("error", fmt.Sprintf("error opening file: %v", err))
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	node, err := parser.ParseFile(bt.fset, filename, nil, parser.AllErrors)
	if err != nil {
		gl.Log("error", fmt.Sprintf("erro ao parsear arquivo: %v", err))
		return nil, fmt.Errorf("erro ao parsear arquivo: %w", err)
	}

	var suggestions []StructBitwiseSuggestion

	ast.Inspect(node, func(n ast.Node) bool {
		st, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		structType, ok := st.Type.(*ast.StructType)
		if !ok {
			return true
		}
		var boolFields []string
		for _, field := range structType.Fields.List {
			if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "bool" {
				for _, name := range field.Names {
					boolFields = append(boolFields, name.Name)
				}
			}
		}
		if len(boolFields) > 0 {
			pos := bt.fset.Position(st.Pos())
			suggestions = append(suggestions, StructBitwiseSuggestion{
				StructName: st.Name.Name,
				BoolFields: boolFields,
				File:       pos.Filename,
				Line:       pos.Line,
			})
		}
		return true
	})
	return suggestions, nil
}

// GenerateTranspiledCode generates optimized bitwise code for a file
func (bt *BitwiseTranspiler) GenerateTranspiledCode(filename string) (string, error) {
	generator := NewBitwiseCodeGenerator()
	return generator.GenerateBitwiseCode(filename)
}

// SugereBitwiseParaArquivo executes analysis and prints suggestions for bitwise conversion
func SugereBitwiseParaArquivo(filename string) error {
	bt := NewBitwiseTranspiler()
	suggestions, err := bt.analyzeFileSugereBitwise(filename)
	if err != nil {
		gl.Log("error", fmt.Sprintf("erro na análise bitwise: %v", err))
		return fmt.Errorf("erro na análise bitwise: %w", err)
	}
	if len(suggestions) == 0 {
		gl.Log("info", fmt.Sprintf("Nenhuma struct com campos bool encontrada em %s\n", filename))
		return nil
	}
	gl.Log("info", "\n--- SUGESTÃO DE CONVERSÃO BITWISE ---\n")
	for _, s := range suggestions {
		gl.Log("info", fmt.Sprintf("Arquivo: %s (linha %d)\nStruct: %s\nCampos bool: %v\nSugestão: Converter para uint64 Flags (bitwise)\n\n", s.File, s.Line, s.StructName, s.BoolFields))
	}
	return nil
}
