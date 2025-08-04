// Package cli provides transpilation commands for GASType
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/faelmori/gastype/internal/transpiler"
	l "github.com/faelmori/logz"
	"github.com/spf13/cobra"
)

// TranspileConfig holds configuration for transpilation operations
type TranspileConfig struct {
	InputPath      string `json:"input_path"`
	OutputPath     string `json:"output_path"`
	Mode           string `json:"mode"`          // "analyze", "transpile", "both"
	OutputFormat   string `json:"output_format"` // "json", "yaml", "text"
	Verbose        bool   `json:"verbose"`
	SecurityLevel  int    `json:"security_level"` // 1=low, 2=medium, 3=high
	PreserveDocs   bool   `json:"preserve_docs"`
	BackupOriginal bool   `json:"backup_original"`
}

// transpileCmd creates the main transpile command
func transpileCmd() *cobra.Command {
	var config TranspileConfig

	cmd := &cobra.Command{
		Use:   "transpile",
		Short: "Transpile Go code to bitwise-optimized equivalent",
		Long: `Transpile traditional Go code to bitwise-optimized equivalent using AST analysis.

This command analyzes Go source code and identifies optimization opportunities:
- Boolean struct fields → Bitwise flags  
- If/else chains → Jump tables
- String literals → Byte arrays (security)
- Configuration structs → Flag systems

The transpiler can operate in different modes:
- analyze: Only analyze and report optimization opportunities
- transpile: Generate transpiled code files
- both: Analyze and generate transpiled code

Examples:
  gastype transpile -i ./src -o ./src_optimized -m transpile
  gastype transpile -i ./config.go -m analyze --format json
  gastype transpile -i ./project -m both --security 3 --verbose`,

		RunE: func(cmd *cobra.Command, args []string) error {
			return runTranspileCommand(&config)
		},
	}

	// Input/Output flags
	cmd.Flags().StringVarP(&config.InputPath, "input", "i", ".",
		"Input directory or file containing Go code to analyze/transpile")
	cmd.Flags().StringVarP(&config.OutputPath, "output", "o", "./gastype_output",
		"Output directory for transpiled code and analysis results")

	// Mode and format flags
	cmd.Flags().StringVarP(&config.Mode, "mode", "m", "analyze",
		"Operation mode: analyze, transpile, or both")
	cmd.Flags().StringVar(&config.OutputFormat, "format", "json",
		"Output format for analysis results: json, yaml, or text")

	// Optimization flags
	cmd.Flags().IntVar(&config.SecurityLevel, "security", 2,
		"Security optimization level (1=low, 2=medium, 3=high)")
	cmd.Flags().BoolVar(&config.PreserveDocs, "preserve-docs", true,
		"Preserve original comments and documentation in transpiled code")
	cmd.Flags().BoolVar(&config.BackupOriginal, "backup", true,
		"Create backup of original files before transpilation")

	// Utility flags
	cmd.Flags().BoolVarP(&config.Verbose, "verbose", "v", false,
		"Show detailed logs of analysis and transpilation process")

	return cmd
}

// runTranspileCommand executes the transpilation process
func runTranspileCommand(config *TranspileConfig) error {
	if config.Verbose {
		l.Info(fmt.Sprintf("🚀 Starting GASType transpilation in mode: %s", config.Mode), nil)
		l.Info(fmt.Sprintf("📁 Input: %s", config.InputPath), nil)
		l.Info(fmt.Sprintf("📁 Output: %s", config.OutputPath), nil)
	}

	// Validate input path
	inputInfo, err := os.Stat(config.InputPath)
	if err != nil {
		return fmt.Errorf("error accessing input path %s: %w", config.InputPath, err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(config.OutputPath, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	// Initialize the bitwise transpiler
	bitwiseTranspiler := transpiler.NewBitwiseTranspiler()

	var results []transpiler.TranspilationResult

	if inputInfo.IsDir() {
		// Analyze entire project
		if config.Verbose {
			l.Info("🔍 Analyzing project directory...", nil)
		}
		results, err = bitwiseTranspiler.AnalyzeProject(config.InputPath)
		if err != nil {
			return fmt.Errorf("error analyzing project: %w", err)
		}
	} else {
		// Analyze single file
		if config.Verbose {
			l.Info("🔍 Analyzing single file...", nil)
		}
		result, err := bitwiseTranspiler.AnalyzeFile(config.InputPath)
		if err != nil {
			return fmt.Errorf("error analyzing file: %w", err)
		}
		if len(result.Optimizations) > 0 {
			results = append(results, *result)
		}
	}

	// Process results based on mode
	switch config.Mode {
	case "analyze":
		return outputAnalysisResults(results, config)
	case "transpile":
		return performTranspilation(results, config)
	case "both":
		if err := outputAnalysisResults(results, config); err != nil {
			return err
		}
		return performTranspilation(results, config)
	default:
		return fmt.Errorf("invalid mode: %s (must be 'analyze', 'transpile', or 'both')", config.Mode)
	}
}

// outputAnalysisResults outputs the analysis results in the specified format
func outputAnalysisResults(results []transpiler.TranspilationResult, config *TranspileConfig) error {
	if len(results) == 0 {
		if config.Verbose {
			l.Info("✅ No optimization opportunities found", nil)
		}
		return nil
	}

	// Create analysis summary
	summary := createAnalysisSummary(results)

	switch config.OutputFormat {
	case "json":
		return outputJSON(results, summary, config)
	case "yaml":
		return outputYAML(results, summary, config)
	case "text":
		return outputText(results, summary, config)
	default:
		return fmt.Errorf("invalid output format: %s", config.OutputFormat)
	}
}

// performTranspilation generates the actual transpiled code
func performTranspilation(results []transpiler.TranspilationResult, config *TranspileConfig) error {
	if config.Verbose {
		l.Info("🔧 Generating transpiled code...", nil)
	}

	// TODO: Implement actual code generation
	// For now, create placeholder files showing what would be generated

	for _, result := range results {
		if err := generateTranspiledFile(result, config); err != nil {
			return fmt.Errorf("error generating transpiled file for %s: %w",
				result.OriginalFile, err)
		}
	}

	if config.Verbose {
		l.Info(fmt.Sprintf("✅ Generated %d transpiled files", len(results)), nil)
	}

	return nil
}

// createAnalysisSummary creates an overall summary of analysis results
func createAnalysisSummary(results []transpiler.TranspilationResult) map[string]interface{} {
	totalOptimizations := 0
	totalBytesSaved := 0
	totalSpeedupFactor := 0.0
	securityFeatures := 0

	for _, result := range results {
		totalOptimizations += len(result.Optimizations)
		for _, opt := range result.Optimizations {
			totalBytesSaved += opt.BytesSaved
			totalSpeedupFactor += opt.SpeedupFactor
		}
		securityFeatures += len(result.SecurityFeatures)
	}

	averageSpeedup := 1.0
	if totalOptimizations > 0 {
		averageSpeedup = totalSpeedupFactor / float64(totalOptimizations)
	}

	return map[string]interface{}{
		"files_analyzed":         len(results),
		"total_optimizations":    totalOptimizations,
		"estimated_bytes_saved":  totalBytesSaved,
		"average_speedup_factor": averageSpeedup,
		"security_features":      securityFeatures,
		"timestamp":              "2025-08-04", // Would use time.Now() in real implementation
	}
}

// outputJSON outputs analysis results in JSON format
func outputJSON(results []transpiler.TranspilationResult, summary map[string]interface{}, config *TranspileConfig) error {
	outputFile := filepath.Join(config.OutputPath, "analysis_results.json")

	output := map[string]interface{}{
		"summary": summary,
		"results": results,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("error writing JSON file: %w", err)
	}

	if config.Verbose {
		l.Info(fmt.Sprintf("📄 Analysis results saved to: %s", outputFile), nil)
	}

	return nil
}

// outputYAML outputs analysis results in YAML format
func outputYAML(results []transpiler.TranspilationResult, summary map[string]interface{}, config *TranspileConfig) error {
	// TODO: Implement YAML output
	return fmt.Errorf("YAML output not yet implemented")
}

// outputText outputs analysis results in human-readable text format
func outputText(results []transpiler.TranspilationResult, summary map[string]interface{}, config *TranspileConfig) error {
	outputFile := filepath.Join(config.OutputPath, "analysis_results.txt")

	var content string
	content += "🚀 GASType Bitwise Analysis Results\n"
	content += "=====================================\n\n"

	// Summary section
	content += fmt.Sprintf("📊 Summary:\n")
	content += fmt.Sprintf("  Files analyzed: %v\n", summary["files_analyzed"])
	content += fmt.Sprintf("  Total optimizations: %v\n", summary["total_optimizations"])
	content += fmt.Sprintf("  Estimated bytes saved: %v\n", summary["estimated_bytes_saved"])
	content += fmt.Sprintf("  Average speedup factor: %.2fx\n", summary["average_speedup_factor"])
	content += fmt.Sprintf("  Security features: %v\n", summary["security_features"])
	content += "\n"

	// Detailed results
	for i, result := range results {
		content += fmt.Sprintf("📁 File %d: %s\n", i+1, result.OriginalFile)
		content += fmt.Sprintf("  Optimizations found: %d\n", len(result.Optimizations))

		for j, opt := range result.Optimizations {
			content += fmt.Sprintf("    %d. %s\n", j+1, opt.Description)
			content += fmt.Sprintf("       Type: %s\n", opt.Type)
			content += fmt.Sprintf("       Location: %s\n", opt.Location)
			content += fmt.Sprintf("       Bytes saved: %d\n", opt.BytesSaved)
			content += fmt.Sprintf("       Speedup: %.2fx\n", opt.SpeedupFactor)
		}

		if len(result.SecurityFeatures) > 0 {
			content += "  Security features:\n"
			for j, feature := range result.SecurityFeatures {
				content += fmt.Sprintf("    %d. %s (%s)\n", j+1, feature.Description, feature.Strength)
			}
		}
		content += "\n"
	}

	if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("error writing text file: %w", err)
	}

	if config.Verbose {
		l.Info(fmt.Sprintf("📄 Analysis results saved to: %s", outputFile), nil)
	}

	return nil
}

// generateTranspiledFile creates a transpiled version of a source file
func generateTranspiledFile(result transpiler.TranspilationResult, config *TranspileConfig) error {
	// For now, create a placeholder showing what optimizations would be applied
	outputFile := filepath.Join(config.OutputPath, filepath.Base(result.TranspiledFile))

	content := fmt.Sprintf(`// Transpiled by GASType - Bitwise Optimization Engine
// Original file: %s
// Optimizations applied: %d

package main

// This file shows what optimizations would be applied:

`, result.OriginalFile, len(result.Optimizations))

	for _, opt := range result.Optimizations {
		content += fmt.Sprintf("// %s: %s\n", opt.Type, opt.Description)
		content += fmt.Sprintf("// Location: %s\n", opt.Location)
		content += fmt.Sprintf("// Performance gain: %.2fx faster\n", opt.SpeedupFactor)
		content += fmt.Sprintf("// Memory saved: %d bytes\n\n", opt.BytesSaved)
	}

	content += "// TODO: Actual transpiled code would be generated here\n"

	if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("error writing transpiled file: %w", err)
	}

	return nil
}

// TranspileCmds returns all transpilation-related commands
func TranspileCmds() []*cobra.Command {
	return []*cobra.Command{
		transpileCmd(),
	}
}
