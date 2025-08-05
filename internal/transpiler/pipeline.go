// Package transpiler provides pipeline functionality for coordinating multiple transpilation passes
package transpiler

import (
	"strings"
)

// DefaultPipeline creates a pipeline with the specified passes
func DefaultPipeline(passes string) []TranspilePass {
	var selected []TranspilePass

	for _, p := range strings.Split(passes, ",") {
		switch strings.TrimSpace(p) {
		case "bool2flags", "bool-to-flags":
			selected = append(selected, &BoolToFlagsPass{})
		case "if2bitwise", "if-to-bitwise":
			selected = append(selected, &IfToBitwisePass{})
		case "assign2bitwise", "assign-to-bitwise":
			selected = append(selected, &AssignToBitwisePass{})
		case "stringobf", "string-obfuscate":
			selected = append(selected, &StringObfuscatePass{})
		case "jumptable", "jump-table":
			selected = append(selected, &JumpTablePass{})
		}
	}

	return selected
}

// RevolutionPipeline returns all available passes for maximum transformation
func RevolutionPipeline() []TranspilePass {
	return []TranspilePass{
		&BoolToFlagsPass{},     // Convert bool fields to bitwise flags
		&IfToBitwisePass{},     // Convert bool conditions to bitwise checks
		&AssignToBitwisePass{}, // Convert bool assignments to bitwise operations
		&StringObfuscatePass{}, // Obfuscate string literals
		&JumpTablePass{},       // Optimize if-chains to jump tables
	}
}

// GetAvailablePasses returns a list of all available pass names
func GetAvailablePasses() []string {
	return []string{
		"bool2flags",
		"if2bitwise",
		"assign2bitwise",
		"stringobf",
		"jumptable",
	}
}
