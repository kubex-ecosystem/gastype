// Package transpiler provides pipeline functionality for coordinating multiple transpilation passes
package transpiler

import (
	"strings"

	"github.com/rafa-mori/gastype/internal/pass"
)

// DefaultPipeline creates a pipeline with the specified passes
func DefaultPipeline(passes string) []TranspilePass {
	var selected []TranspilePass

	for _, p := range strings.Split(passes, ",") {
		switch strings.TrimSpace(p) {
		case "bool2flags", "bool-to-flags":
			selected = append(selected, pass.NewBoolToFlagsPass())
		case "if2bitwise", "if-to-bitwise":
			selected = append(selected, pass.NewIfToBitwisePass())
		case "assign2bitwise", "assign-to-bitwise":
			selected = append(selected, pass.NewAssignToBitwisePass())
		case "field2bitwise", "field-to-bitwise":
			selected = append(selected, pass.NewFieldAccessToBitwisePass())
		case "stringobf", "string-obfuscate":
			selected = append(selected, pass.NewStringObfuscatePass())
		case "jumptable", "jump-table":
			selected = append(selected, pass.NewJumpTablePass())
		}
	}

	return selected
}

// ProcessPipeline returns all available passes for maximum transformation
func ProcessPipeline() []TranspilePass {
	return []TranspilePass{
		pass.NewBoolToFlagsPass(),          // Convert bool fields to bitwise flags
		pass.NewIfToBitwisePass(),          // Convert bool conditions to bitwise checks
		pass.NewAssignToBitwisePass(),      // Convert bool assignments to bitwise operations
		pass.NewFieldAccessToBitwisePass(), // 🚀: Convert field access to bitwise checks
		pass.NewStringObfuscatePass(),      // Obfuscate string literals
		pass.NewJumpTablePass(),            // Optimize if-chains to jump tables
	}
}

// GetAvailablePasses returns a list of all available pass names
func GetAvailablePasses() []string {
	return []string{
		"bool2flags",
		"if2bitwise",
		"assign2bitwise",
		"field2bitwise",
		"stringobf",
		"jumptable",
	}
}
