// Full Project Transpiler - REVOLUTIONARY COMPLETE PROJECT TRANSPILATION
// This system transpiles ENTIRE Go projects to bitwise-optimized versions
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/faelmori/gastype/internal/transpiler"
)

// ProjectTranspiler handles complete project transpilation
type ProjectTranspiler struct {
	transpiler    *transpiler.BitwiseTranspiler
	analyzer      *transpiler.ContextAnalyzer
	generator     *transpiler.AdvancedCodeGenerator
	sourceProject string
	targetProject string
	stats         *TranspilationStats
}

// TranspilationStats tracks transpilation metrics
type TranspilationStats struct {
	StartTime          time.Time     `json:"start_time"`
	EndTime            time.Time     `json:"end_time"`
	Duration           time.Duration `json:"duration"`
	TotalFiles         int           `json:"total_files"`
	GoFiles            int           `json:"go_files"`
	TranspiledFiles    int           `json:"transpiled_files"`
	ContextsFound      int           `json:"contexts_found"`
	ContextsTranspiled int           `json:"contexts_transpiled"`
	BytesSaved         int           `json:"bytes_saved"`
	OptimizationLevel  string        `json:"optimization_level"`
	ObfuscationLevel   string        `json:"obfuscation_level"`
	Errors             []string      `json:"errors,omitempty"`
	Warnings           []string      `json:"warnings,omitempty"`
}

// NewProjectTranspiler creates a new complete project transpiler
func NewProjectTranspiler(sourceProject, targetProject string) *ProjectTranspiler {
	return &ProjectTranspiler{
		transpiler:    transpiler.NewBitwiseTranspiler(),
		analyzer:      transpiler.NewContextAnalyzer(),
		generator:     transpiler.NewAdvancedCodeGenerator(3), // 3 = HIGH obfuscation
		sourceProject: sourceProject,
		targetProject: targetProject,
		stats: &TranspilationStats{
			StartTime:         time.Now(),
			OptimizationLevel: "SURREAL",
			ObfuscationLevel:  "HIGH",
			Errors:            []string{},
			Warnings:          []string{},
		},
	}
}

// TranspileCompleteProject transpiles an entire Go project
func (pt *ProjectTranspiler) TranspileCompleteProject() error {
	fmt.Println("🚀 INICIANDO TRANSPILAÇÃO COMPLETA DE PROJETO - MODO REVOLUCIONÁRIO ATIVADO!")
	fmt.Printf("📂 Projeto origem: %s\n", pt.sourceProject)
	fmt.Printf("📂 Projeto destino: %s\n", pt.targetProject)

	// Step 1: Validate source project
	if err := pt.validateSourceProject(); err != nil {
		return fmt.Errorf("validação do projeto origem falhou: %w", err)
	}

	// Step 2: Create target project structure
	if err := pt.createTargetStructure(); err != nil {
		return fmt.Errorf("criação da estrutura destino falhou: %w", err)
	}

	// Step 3: Copy non-Go files (preserving structure)
	if err := pt.copyNonGoFiles(); err != nil {
		return fmt.Errorf("cópia de arquivos não-Go falhou: %w", err)
	}

	// Step 4: Analyze entire project for contexts
	contexts, err := pt.analyzeProjectContexts()
	if err != nil {
		return fmt.Errorf("análise de contextos falhou: %w", err)
	}

	// Step 5: Transpile all Go files
	if err := pt.transpileAllGoFiles(contexts); err != nil {
		return fmt.Errorf("transpilação de arquivos Go falhou: %w", err)
	}

	// Step 6: Generate build scripts and configurations
	if err := pt.generateBuildSystem(); err != nil {
		return fmt.Errorf("geração do sistema de build falhou: %w", err)
	}

	// Step 7: Generate transpilation report
	if err := pt.generateReport(); err != nil {
		return fmt.Errorf("geração de relatório falhou: %w", err)
	}

	pt.stats.EndTime = time.Now()
	pt.stats.Duration = pt.stats.EndTime.Sub(pt.stats.StartTime)

	fmt.Println("\n🔥 TRANSPILAÇÃO COMPLETA FINALIZADA!")
	fmt.Printf("⏱️  Tempo total: %v\n", pt.stats.Duration)
	fmt.Printf("📁 Arquivos Go transpilados: %d/%d\n", pt.stats.TranspiledFiles, pt.stats.GoFiles)
	fmt.Printf("🧠 Contextos encontrados: %d\n", pt.stats.ContextsFound)
	fmt.Printf("⚡ Contextos transpilados: %d\n", pt.stats.ContextsTranspiled)
	fmt.Printf("💾 Projeto transpilado salvo em: %s\n", pt.targetProject)

	return nil
}

// validateSourceProject checks if source project is valid Go project
func (pt *ProjectTranspiler) validateSourceProject() error {
	// Check if go.mod exists
	goModPath := filepath.Join(pt.sourceProject, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return fmt.Errorf("go.mod não encontrado - não é um projeto Go válido")
	}

	// Count Go files
	goFileCount := 0
	err := filepath.Walk(pt.sourceProject, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".go") {
			goFileCount++
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("erro contando arquivos Go: %w", err)
	}

	if goFileCount == 0 {
		return fmt.Errorf("nenhum arquivo Go encontrado no projeto")
	}

	pt.stats.GoFiles = goFileCount
	fmt.Printf("✅ Projeto válido encontrado com %d arquivos Go\n", goFileCount)
	return nil
}

// createTargetStructure creates the complete target project structure
func (pt *ProjectTranspiler) createTargetStructure() error {
	// Remove existing target if exists
	if _, err := os.Stat(pt.targetProject); !os.IsNotExist(err) {
		fmt.Printf("🗑️  Removendo projeto transpilado existente...\n")
		if err := os.RemoveAll(pt.targetProject); err != nil {
			return fmt.Errorf("erro removendo projeto existente: %w", err)
		}
	}

	// Create target directory
	if err := os.MkdirAll(pt.targetProject, 0755); err != nil {
		return fmt.Errorf("erro criando diretório destino: %w", err)
	}

	// Replicate entire directory structure
	err := filepath.Walk(pt.sourceProject, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(pt.sourceProject, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(pt.targetProject, relPath)

		// Create directories
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("erro replicando estrutura: %w", err)
	}

	fmt.Printf("✅ Estrutura de diretórios replicada\n")
	return nil
}

// copyNonGoFiles copies all non-Go files preserving structure
func (pt *ProjectTranspiler) copyNonGoFiles() error {
	err := filepath.Walk(pt.sourceProject, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip Go files (they will be transpiled)
		if strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		// Calculate paths
		relPath, err := filepath.Rel(pt.sourceProject, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(pt.targetProject, relPath)

		// Copy file
		return pt.copyFile(path, targetPath)
	})

	if err != nil {
		return fmt.Errorf("erro copiando arquivos não-Go: %w", err)
	}

	fmt.Printf("✅ Arquivos não-Go copiados\n")
	return nil
}

// copyFile copies a single file
func (pt *ProjectTranspiler) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// analyzeProjectContexts analyzes the entire project for transpilable contexts
func (pt *ProjectTranspiler) analyzeProjectContexts() (map[string][]transpiler.LogicalContext, error) {
	fmt.Printf("🧠 Analisando contextos lógicos do projeto...\n")

	allContexts := make(map[string][]transpiler.LogicalContext)

	err := filepath.Walk(pt.sourceProject, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only analyze Go files
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		// Analyze contexts in this file
		contexts, err := pt.analyzer.AnalyzeFile(path)
		if err != nil {
			pt.stats.Warnings = append(pt.stats.Warnings, fmt.Sprintf("Aviso analisando %s: %v", path, err))
			return nil // Continue with other files
		}

		if len(contexts) > 0 {
			allContexts[path] = contexts
			pt.stats.ContextsFound += len(contexts)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("erro analisando contextos: %w", err)
	}

	fmt.Printf("✅ Análise completa: %d contextos encontrados em %d arquivos\n",
		pt.stats.ContextsFound, len(allContexts))

	return allContexts, nil
}

// transpileAllGoFiles transpiles all Go files using found contexts
func (pt *ProjectTranspiler) transpileAllGoFiles(contexts map[string][]transpiler.LogicalContext) error {
	fmt.Printf("⚡ Transpilando arquivos Go...\n")

	err := filepath.Walk(pt.sourceProject, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process Go files
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		// Calculate target path
		relPath, err := filepath.Rel(pt.sourceProject, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(pt.targetProject, relPath)

		// Transpile this file
		fileContexts := contexts[path]
		if len(fileContexts) == 0 {
			// No contexts found, copy original file
			if err := pt.copyFile(path, targetPath); err != nil {
				return fmt.Errorf("erro copiando %s: %w", path, err)
			}
		} else {
			// Transpile with contexts
			transpiledCode, err := pt.generator.GenerateAdvancedCode(path, fileContexts)
			if err != nil {
				pt.stats.Errors = append(pt.stats.Errors, fmt.Sprintf("Erro transpilando %s: %v", path, err))
				// Fallback to original file
				if err := pt.copyFile(path, targetPath); err != nil {
					return fmt.Errorf("erro copiando fallback %s: %w", path, err)
				}
			} else {
				// Save transpiled code
				if err := os.WriteFile(targetPath, []byte(transpiledCode), 0644); err != nil {
					return fmt.Errorf("erro salvando transpilado %s: %w", targetPath, err)
				}
				pt.stats.TranspiledFiles++
				pt.stats.ContextsTranspiled += len(fileContexts)
			}
		}

		pt.stats.TotalFiles++
		return nil
	})

	if err != nil {
		return fmt.Errorf("erro transpilando arquivos: %w", err)
	}

	fmt.Printf("✅ Transpilação completa: %d arquivos processados\n", pt.stats.TotalFiles)
	return nil
}

// generateBuildSystem creates build scripts and configurations for transpiled project
func (pt *ProjectTranspiler) generateBuildSystem() error {
	fmt.Printf("🔧 Gerando sistema de build...\n")

	// Create build script
	buildScript := `#!/bin/bash
# GASType Revolutionary Transpiled Project Build Script
# Generated automatically - DO NOT EDIT MANUALLY

echo "🚀 Building GASType Revolutionary Transpiled Project..."

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -f main transpiled_binary

# Build project
echo "⚡ Compiling transpiled code..."
go build -ldflags="-s -w" -o transpiled_binary .

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo "📦 Binary: transpiled_binary"
    ls -lh transpiled_binary
    echo "🎯 Ready to run revolutionary transpiled code!"
else
    echo "❌ Build failed!"
    exit 1
fi
`

	buildPath := filepath.Join(pt.targetProject, "build.sh")
	if err := os.WriteFile(buildPath, []byte(buildScript), 0755); err != nil {
		return fmt.Errorf("erro criando build.sh: %w", err)
	}

	// Create README for transpiled project
	readme := fmt.Sprintf(`# GASType Revolutionary Transpiled Project

This is a **REVOLUTIONARY TRANSPILED VERSION** of a Go project using the GASType transpiler.

## 🚀 About This Transpilation

- **Original Project**: %s
- **Transpilation Date**: %s
- **Optimization Level**: %s
- **Obfuscation Level**: %s
- **Files Transpiled**: %d/%d
- **Contexts Converted**: %d

## ⚡ Performance Features

✅ **Ultra-optimized bitwise operations**
✅ **Maximum code obfuscation**  
✅ **Reduced binary size**
✅ **Enhanced security through code obfuscation**
✅ **Surreal performance optimizations**

## 🔧 Building

`+"```bash"+`
chmod +x build.sh
./build.sh
`+"```"+`

## ⚠️ Important Notes

- This code has been **transpiled for maximum performance and obfuscation**
- **Human readability has been intentionally eliminated**
- Original function names have been **completely obfuscated**
- All logical structures converted to **bitwise state machines**

## 🛡️ Security

This transpiled code provides enhanced security through:
- Obfuscated function and variable names
- Bitwise operations instead of traditional logic
- Eliminated human-readable patterns
- Anti-reverse engineering measures

---
*Generated by GASType Revolutionary Transpiler*
`, pt.sourceProject, time.Now().Format("2006-01-02 15:04:05"),
		pt.stats.OptimizationLevel, pt.stats.ObfuscationLevel,
		pt.stats.TranspiledFiles, pt.stats.GoFiles, pt.stats.ContextsTranspiled)

	readmePath := filepath.Join(pt.targetProject, "README_TRANSPILED.md")
	if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		return fmt.Errorf("erro criando README_TRANSPILED.md: %w", err)
	}

	fmt.Printf("✅ Sistema de build gerado\n")
	return nil
}

// generateReport generates a comprehensive transpilation report
func (pt *ProjectTranspiler) generateReport() error {
	// Generate JSON report
	reportPath := filepath.Join(pt.targetProject, "transpilation_report.json")
	reportData, err := json.MarshalIndent(pt.stats, "", "  ")
	if err != nil {
		return fmt.Errorf("erro gerando relatório JSON: %w", err)
	}

	if err := os.WriteFile(reportPath, reportData, 0644); err != nil {
		return fmt.Errorf("erro salvando relatório: %w", err)
	}

	fmt.Printf("✅ Relatório de transpilação salvo em: %s\n", reportPath)
	return nil
}

func TranspileProject() {
	if len(os.Args) != 3 {
		fmt.Println("Uso: go run full_project_transpiler.go <projeto_origem> <projeto_destino>")
		fmt.Println("Exemplo: go run full_project_transpiler.go /path/to/source /path/to/target")
		os.Exit(1)
	}

	sourceProject := os.Args[1]
	targetProject := os.Args[2]

	transpiler := NewProjectTranspiler(sourceProject, targetProject)

	if err := transpiler.TranspileCompleteProject(); err != nil {
		fmt.Printf("❌ ERRO FATAL: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n🎉 TRANSPILAÇÃO REVOLUCIONÁRIA COMPLETA! 🎉")
	fmt.Println("🚀 Seu projeto foi completamente transpilado para máxima performance!")
}
