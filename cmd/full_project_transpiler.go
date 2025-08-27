// Full Project Transpiler - COMPLETE PROJECT TRANSPILATION
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

	transpiler "github.com/rafa-mori/gastype/internal/engine"

	gl "github.com/rafa-mori/gastype/internal/module/logger"
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
	gl.Log("info", "🚀 STARTING COMPLETE PROJECT TRANSPILATION - MODE ACTIVATED!")
	gl.Log("info", fmt.Sprintf("📂 Source project: %s", pt.sourceProject))
	gl.Log("info", fmt.Sprintf("📂 Target project: %s", pt.targetProject))

	// Step 1: Validate source project
	if err := pt.validateSourceProject(); err != nil {
		gl.Log("error", fmt.Sprintf("source project validation failed: %v", err))
		return fmt.Errorf("source project validation failed: %w", err)
	}

	// Step 2: Create target project structure
	if err := pt.createTargetStructure(); err != nil {
		gl.Log("error", fmt.Sprintf("target structure creation failed: %v", err))
		return fmt.Errorf("target structure creation failed: %w", err)
	}

	// Step 3: Copy non-Go files (preserving structure)
	if err := pt.copyNonGoFiles(); err != nil {
		gl.Log("error", fmt.Sprintf("non-Go files copy failed: %v", err))
		return fmt.Errorf("non-Go files copy failed: %w", err)
	}

	// Step 4: Analyze entire project for contexts
	contexts, err := pt.analyzeProjectContexts()
	if err != nil {
		gl.Log("error", fmt.Sprintf("context analysis failed: %v", err))
		return fmt.Errorf("context analysis failed: %w", err)
	}

	// Step 5: Transpile all Go files
	if err := pt.transpileAllGoFiles(contexts); err != nil {
		gl.Log("error", fmt.Sprintf("go files transpilation failed: %v", err))
		return fmt.Errorf("go files transpilation failed: %w", err)
	}

	// Step 6: Generate build scripts and configurations
	if err := pt.generateBuildSystem(); err != nil {
		gl.Log("error", fmt.Sprintf("build system generation failed: %v", err))
		return fmt.Errorf("build system generation failed: %w", err)
	}

	// Step 7: Generate transpilation report
	if err := pt.generateReport(); err != nil {
		gl.Log("error", fmt.Sprintf("report generation failed: %v", err))
		return fmt.Errorf("report generation failed: %w", err)
	}

	pt.stats.EndTime = time.Now()
	pt.stats.Duration = pt.stats.EndTime.Sub(pt.stats.StartTime)

	gl.Log("info", "🔥 COMPLETE TRANSPILATION FINISHED!")
	gl.Log("info", fmt.Sprintf("⏱️  Total time: %v", pt.stats.Duration))
	gl.Log("info", fmt.Sprintf("📁 Go files transpiled: %d/%d", pt.stats.TranspiledFiles, pt.stats.GoFiles))
	gl.Log("info", fmt.Sprintf("🧠 Contexts found: %d", pt.stats.ContextsFound))
	gl.Log("info", fmt.Sprintf("⚡ Contexts transpiled: %d", pt.stats.ContextsTranspiled))
	gl.Log("info", fmt.Sprintf("💾 Transpiled project saved at: %s", pt.targetProject))

	return nil
}

// validateSourceProject checks if source project is valid Go project
func (pt *ProjectTranspiler) validateSourceProject() error {
	// Check if go.mod exists
	goModPath := filepath.Join(pt.sourceProject, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		gl.Log("error", "go.mod not found - not a valid Go project")
		return fmt.Errorf("go.mod not found - not a valid Go project")
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
		gl.Log("error", fmt.Sprintf("error counting Go files: %v", err))
		return fmt.Errorf("error counting Go files: %w", err)
	}

	if goFileCount == 0 {
		gl.Log("error", "no Go files found in project")
		return fmt.Errorf("no Go files found in project")
	}

	pt.stats.GoFiles = goFileCount
	gl.Log("info", fmt.Sprintf("✅ Valid project found with %d Go files", goFileCount))
	return nil
}

// createTargetStructure creates the complete target project structure
func (pt *ProjectTranspiler) createTargetStructure() error {
	// Remove existing target if exists
	if _, err := os.Stat(pt.targetProject); !os.IsNotExist(err) {
		gl.Log("info", "🗑️  Removing existing transpiled project...")
		if err := os.RemoveAll(pt.targetProject); err != nil {
			gl.Log("error", fmt.Sprintf("error removing existing project: %v", err))
			return fmt.Errorf("error removing existing project: %w", err)
		}
	}

	// Create target directory
	if err := os.MkdirAll(pt.targetProject, 0755); err != nil {
		gl.Log("error", fmt.Sprintf("error creating target directory: %v", err))
		return fmt.Errorf("error creating target directory: %w", err)
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
		gl.Log("error", fmt.Sprintf("error replicating structure: %v", err))
		return fmt.Errorf("error replicating structure: %w", err)
	}

	gl.Log("info", "✅ Directory structure replicated")
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
		gl.Log("error", fmt.Sprintf("error copying non-Go files: %v", err))
		return fmt.Errorf("error copying non-Go files: %w", err)
	}

	gl.Log("info", "✅ Non-Go files copied")
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
	gl.Log("info", "🧠 Analyzing project logical contexts...")

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
			pt.stats.Warnings = append(pt.stats.Warnings, fmt.Sprintf("Warning analyzing %s: %v", path, err))
			return nil // Continue with other files
		}

		if len(contexts) > 0 {
			allContexts[path] = contexts
			pt.stats.ContextsFound += len(contexts)
		}

		return nil
	})

	if err != nil {
		gl.Log("error", fmt.Sprintf("error analyzing contexts: %s", err))
		return nil, fmt.Errorf("error analyzing contexts: %w", err)
	}

	gl.Log("info", fmt.Sprintf("✅ Complete analysis: %d contexts found in %d files",
		pt.stats.ContextsFound, len(allContexts)))

	return allContexts, nil
}

// transpileAllGoFiles transpiles all Go files using found contexts
func (pt *ProjectTranspiler) transpileAllGoFiles(contexts map[string][]transpiler.LogicalContext) error {
	gl.Log("info", "⚡ Transpiling Go files...")

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
				gl.Log("error", fmt.Sprintf("error copying %s: %s", path, err))
				return fmt.Errorf("error copying %s: %w", path, err)
			}
		} else {
			// Transpile with contexts
			transpiledCode, err := pt.generator.GenerateAdvancedCode(path, fileContexts)
			if err != nil {
				pt.stats.Errors = append(pt.stats.Errors, fmt.Sprintf("Error transpiling %s: %v", path, err))
				// Fallback to original file
				if err := pt.copyFile(path, targetPath); err != nil {
					gl.Log("error", fmt.Sprintf("error copying fallback %s: %s", path, err))
					return fmt.Errorf("error copying fallback %s: %w", path, err)
				}
			} else {
				// Save transpiled code
				if err := os.WriteFile(targetPath, []byte(transpiledCode), 0644); err != nil {
					gl.Log("error", fmt.Sprintf("error saving transpiled %s: %s", targetPath, err))
					return fmt.Errorf("error saving transpiled %s: %w", targetPath, err)
				}
				pt.stats.TranspiledFiles++
				pt.stats.ContextsTranspiled += len(fileContexts)
			}
		}

		pt.stats.TotalFiles++
		return nil
	})

	if err != nil {
		gl.Log("error", fmt.Sprintf("error transpiling files: %s", err))
		return fmt.Errorf("error transpiling files: %w", err)
	}

	gl.Log("info", fmt.Sprintf("✅ Complete transpilation: %d files processed", pt.stats.TotalFiles))
	return nil
}

// generateBuildSystem creates build scripts and configurations for transpiled project
func (pt *ProjectTranspiler) generateBuildSystem() error {
	gl.Log("info", "🔧 Generating build system...")

	// Create build script
	buildScript := `#!/bin/bash
# GASType  Transpiled Project Build Script
# Generated automatically - DO NOT EDIT MANUALLY

echo "🚀 Building GASType  Transpiled Project..."

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
    echo "🎯 Ready to run transpiled code!"
else
    echo "❌ Build failed!"
    exit 1
fi
`

	buildPath := filepath.Join(pt.targetProject, "build.sh")
	if err := os.WriteFile(buildPath, []byte(buildScript), 0755); err != nil {
		gl.Log("error", fmt.Sprintf("error creating build.sh: %v", err))
		return fmt.Errorf("error creating build.sh: %w", err)
	}

	// Create README for transpiled project
	readme := fmt.Sprintf(`# GASType  Transpiled Project

This is a **TRANSPILED VERSION** of a Go project using the GASType transpiler.

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
*Generated by GASType  Transpiler*
`, pt.sourceProject, time.Now().Format("2006-01-02 15:04:05"),
		pt.stats.OptimizationLevel, pt.stats.ObfuscationLevel,
		pt.stats.TranspiledFiles, pt.stats.GoFiles, pt.stats.ContextsTranspiled)

	readmePath := filepath.Join(pt.targetProject, "README_TRANSPILED.md")
	if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		gl.Log("error", fmt.Sprintf("error creating README_TRANSPILED.md: %v", err))
		return fmt.Errorf("error creating README_TRANSPILED.md: %w", err)
	}

	gl.Log("info", "✅ Build system generated")
	return nil
}

// generateReport generates a comprehensive transpilation report
func (pt *ProjectTranspiler) generateReport() error {
	// Generate JSON report
	reportPath := filepath.Join(pt.targetProject, "transpilation_report.json")
	reportData, err := json.MarshalIndent(pt.stats, "", "  ")
	if err != nil {
		gl.Log("error", fmt.Sprintf("error generating JSON report: %v", err))
		return fmt.Errorf("error generating JSON report: %w", err)
	}

	if err := os.WriteFile(reportPath, reportData, 0644); err != nil {
		gl.Log("error", fmt.Sprintf("error saving report: %v", err))
		return fmt.Errorf("error saving report: %w", err)
	}

	gl.Log("info", fmt.Sprintf("✅ Transpilation report saved at: %s", reportPath))
	return nil
}
func TranspileProject() {
	if len(os.Args) != 3 {
		gl.Log("error", "Usage: go run full_project_transpiler.go <source_project> <target_project>")
		gl.Log("fatal", "Example: go run full_project_transpiler.go /path/to/source /path/to/target")
	}

	sourceProject := os.Args[1]
	targetProject := os.Args[2]

	transpiler := NewProjectTranspiler(sourceProject, targetProject)

	if err := transpiler.TranspileCompleteProject(); err != nil {
		gl.Log("fatal", err)
	}

	gl.Log("info", "🎉 TRANSPILATION COMPLETE! 🎉")
	gl.Log("info", "🚀 Your project has been completely transpiled for maximum performance!")
}
