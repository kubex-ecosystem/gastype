// Package transpiler implementa pipeline de transpilação e sugestões bitwise
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

// StructBitwiseSuggestion representa sugestão de conversão de campos bool para flags
type StructBitwiseSuggestion struct {
	StructName string
	BoolFields []string
	File       string
	Line       int
}

// BitwiseTranspiler é a estrutura principal para transpilação bitwise
type BitwiseTranspiler struct {
	fset *token.FileSet
}

// NewBitwiseTranspiler cria um novo transpiler bitwise
func NewBitwiseTranspiler() *BitwiseTranspiler {
	return &BitwiseTranspiler{
		fset: token.NewFileSet(),
	}
}

// TranspilationResult representa o resultado de uma análise/transpilação
type TranspilationResult struct {
	OriginalFile     string            `json:"original_file"`
	TranspiledFile   string            `json:"transpiled_file"`
	Optimizations    []Optimization    `json:"optimizations"`
	SecurityFeatures []SecurityFeature `json:"security_features"`
}

// Optimization representa uma otimização encontrada
type Optimization struct {
	Type          string  `json:"type"`
	Description   string  `json:"description"`
	Location      string  `json:"location"`
	BytesSaved    int     `json:"bytes_saved"`
	SpeedupFactor float64 `json:"speedup_factor"`
}

// SecurityFeature representa um recurso de segurança aplicado
type SecurityFeature struct {
	Description string `json:"description"`
	Strength    string `json:"strength"`
}

// AnalyzeFile analisa um arquivo único e retorna resultados de otimização
func (bt *BitwiseTranspiler) AnalyzeFile(filename string) (*TranspilationResult, error) {
	suggestions, err := bt.analyzeFileSugereBitwise(filename)
	if err != nil {
		gl.Log("error", fmt.Sprintf("erro na análise bitwise: %w", err))
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

// AnalyzeProject analisa um projeto inteiro (diretório)
func (bt *BitwiseTranspiler) AnalyzeProject(projectDir string) ([]TranspilationResult, error) {
	var results []TranspilationResult

	err := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Analisar apenas arquivos .go
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		result, err := bt.AnalyzeFile(path)
		if err != nil {
			gl.Log("error", fmt.Sprintf("erro analisando %s: %w", path, err))
			return fmt.Errorf("erro analisando %s: %w", path, err)
		}

		// Apenas adicionar se há otimizações
		if len(result.Optimizations) > 0 {
			results = append(results, *result)
		}

		return nil
	})

	if err != nil {
		gl.Log("error", fmt.Sprintf("erro percorrendo projeto: %w", err))
		return nil, fmt.Errorf("erro percorrendo projeto: %w", err)
	}

	return results, nil
}

// analyzeFileSugereBitwise analisa um arquivo Go e retorna structs que podem ser convertidos para bitwise
func (bt *BitwiseTranspiler) analyzeFileSugereBitwise(filename string) ([]StructBitwiseSuggestion, error) {
	file, err := os.Open(filename)
	if err != nil {
		gl.Log("error", fmt.Sprintf("erro ao abrir arquivo: %w", err))
		return nil, fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	node, err := parser.ParseFile(bt.fset, filename, nil, parser.AllErrors)
	if err != nil {
		gl.Log("error", fmt.Sprintf("erro ao parsear arquivo: %w", err))
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

// GenerateTranspiledCode gera código bitwise otimizado para um arquivo
func (bt *BitwiseTranspiler) GenerateTranspiledCode(filename string) (string, error) {
	generator := NewBitwiseCodeGenerator()
	return generator.GenerateBitwiseCode(filename)
}

// SugereBitwiseParaArquivo executa análise e imprime sugestões para conversão bitwise
func SugereBitwiseParaArquivo(filename string) error {
	bt := NewBitwiseTranspiler()
	suggestions, err := bt.analyzeFileSugereBitwise(filename)
	if err != nil {
		gl.Log("error", fmt.Sprintf("erro na análise bitwise: %w", err))
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
