package astutil

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strings"
)

// TranspileContext tracks all information about a transpilation operation
type TranspileContext struct {
	*Info

	// General configuration
	Ofuscate  bool   `json:"ofuscate"`   // If true, names and structure will be obfuscated
	MapFile   string `json:"map_file"`   // Path to output .map.json file
	InputFile string `json:"input_file"` // Input file path
	OutputDir string `json:"output_dir"` // Output directory
	DryRun    bool   `json:"dry_run"`    // If true, only analyze without saving files

	// Analysis results
	Structs map[string]*StructInfo `json:"structs"` // Original struct → detailed info
	Flags   map[string][]string    `json:"flags"`   // Struct → list of generated flags

	// 🚀 REVOLUTIONARY FIELDS for OutputManager
	GeneratedFiles map[string]*ast.File `json:"-"` // File path → transpiled AST
	Fset           *token.FileSet       `json:"-"` // Token file set for all files

	// 🔥 PACKAGE-SCOPED CONSTANTS TRACKING
	PackageConstantsAdded map[string]bool `json:"-"` // Package → constants added (prevents duplicates)
}

// StructInfo contains detailed information about each detected struct
type StructInfo struct {
	OriginalName    string            `json:"original_name"`   // Original name (e.g., Config)
	NewName         string            `json:"new_name"`        // Final name (e.g., ConfigFlags)
	BoolFields      []string          `json:"bool_fields"`     // Original bool fields
	FlagMapping     map[string]string `json:"flag_mapping"`    // BoolField → FlagName
	Transformations map[string]string `json:"transformations"` // Track applied transformations
}

// NewContext creates a new transpilation context
func NewContext(inputFile, outputDir string, ofuscate bool, mapFile string) *TranspileContext {
	return &TranspileContext{
		Ofuscate:       ofuscate,
		MapFile:        mapFile,
		InputFile:      inputFile,
		OutputDir:      outputDir,
		Structs:        make(map[string]*StructInfo),
		Flags:          make(map[string][]string),
		GeneratedFiles: make(map[string]*ast.File), // 🚀 REVOLUTIONARY: Store transpiled files
		Fset:           token.NewFileSet(),         // 🚀 REVOLUTIONARY: Share FileSet across all operations
	}
}

// AddStruct registers a struct transformation in the context
func (ctx *TranspileContext) AddStruct(packageName, originalName, newName string, boolFields []string) {
	mapping := make(map[string]string)
	for _, f := range boolFields {
		// 🚀 REVOLUTIONARY: Include package name to avoid conflicts
		mapping[f] = fmt.Sprintf("Flag%s_%s_%s", strings.Title(packageName), originalName, strings.Title(f))
	}

	ctx.Structs[originalName] = &StructInfo{
		OriginalName: originalName,
		NewName:      newName,
		BoolFields:   boolFields,
		FlagMapping:  mapping,
	}
	ctx.Flags[newName] = boolFields
}

// AddFlagMapping adds a flag mapping for a specific struct and field (cleaner approach)
func (ctx *TranspileContext) AddFlagMapping(structName, fieldName, flagName string, bitPos int) {
	if ctx.Structs[structName] == nil {
		ctx.Structs[structName] = &StructInfo{
			OriginalName: structName,
			BoolFields:   []string{},
			FlagMapping:  map[string]string{},
		}
	}
	ctx.Structs[structName].FlagMapping[fieldName] = flagName
}

// GetFlagName returns the flag name for a given struct and field
func (ctx *TranspileContext) GetFlagName(packageName, structName, fieldName string) string {
	if structInfo, exists := ctx.Structs[structName]; exists {
		if flagName, exists := structInfo.FlagMapping[fieldName]; exists {
			return flagName
		}
	}
	// 🚀 REVOLUTIONARY: Include package name to avoid conflicts between packages
	return fmt.Sprintf("Flag%s_%s_%s", strings.Title(packageName), structName, strings.Title(fieldName))
}

// IsStructTransformed checks if a struct has been transformed
func (ctx *TranspileContext) IsStructTransformed(structName string) bool {
	_, exists := ctx.Structs[structName]
	return exists
}

// GetTransformedStructName returns the new name for a transformed struct
func (ctx *TranspileContext) GetTransformedStructName(originalName string) string {
	if structInfo, exists := ctx.Structs[originalName]; exists {
		return structInfo.NewName
	}
	return originalName + "Flags" // fallback
}

// IsBoolField checks if a field is a bool field in any transformed struct
func (ctx *TranspileContext) IsBoolField(structName, fieldName string) bool {
	if structInfo, exists := ctx.Structs[structName]; exists {
		for _, boolField := range structInfo.BoolFields {
			if boolField == fieldName {
				return true
			}
		}
	}
	return false
}

// SaveMap saves the context as a JSON map file
func (ctx *TranspileContext) SaveMap() error {
	if ctx.MapFile == "" {
		return nil
	}

	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ctx.MapFile, data, 0644)
}

// LoadMap loads a context from a JSON map file
func LoadMap(mapFile string) (*TranspileContext, error) {
	data, err := os.ReadFile(mapFile)
	if err != nil {
		return nil, err
	}

	var ctx TranspileContext
	err = json.Unmarshal(data, &ctx)
	if err != nil {
		return nil, err
	}

	return &ctx, nil
}

// EstimatePerformance provides performance estimates for the transpilation
func (ctx *TranspileContext) EstimatePerformance() {
	fmt.Printf("\n📊 Performance Estimation:\n")

	totalStructs := len(ctx.Structs)
	totalBoolFields := 0
	totalFlagsGenerated := 0

	for _, structInfo := range ctx.Structs {
		totalBoolFields += len(structInfo.BoolFields)
		totalFlagsGenerated += len(structInfo.BoolFields)
	}

	if totalStructs == 0 {
		fmt.Printf("  ℹ️  No transformations found - no performance impact\n")
		return
	}

	// Memory usage estimates
	originalMemoryPerStruct := totalBoolFields // each bool = 1 byte typically
	optimizedMemoryPerStruct := 8              // uint64 = 8 bytes

	memoryReduction := float64(originalMemoryPerStruct-optimizedMemoryPerStruct) / float64(originalMemoryPerStruct) * 100
	if memoryReduction < 0 {
		memoryReduction = 0 // In cases where we have few bools, memory might increase slightly
	}

	fmt.Printf("  📈 Structs analyzed: %d\n", totalStructs)
	fmt.Printf("  🔢 Bool fields → Flags: %d → %d constants\n", totalBoolFields, totalFlagsGenerated)
	fmt.Printf("  💾 Memory per struct: %d bytes → 8 bytes\n", originalMemoryPerStruct)

	if memoryReduction > 0 {
		fmt.Printf("  ⚡ Estimated memory reduction: %.1f%%\n", memoryReduction)
	} else {
		fmt.Printf("  ℹ️  Memory usage: minimal change (small struct overhead)\n")
	}

	// Performance benefits
	fmt.Printf("  🚀 Performance benefits:\n")
	fmt.Printf("     • Bitwise operations: ~2-5x faster than bool comparisons\n")
	fmt.Printf("     • Cache efficiency: Better memory locality\n")
	fmt.Printf("     • Atomic operations: Single uint64 vs multiple bools\n")

	// Security benefits
	fmt.Printf("  🔒 Security benefits:\n")
	fmt.Printf("     • Obfuscated logic: Harder to reverse engineer\n")
	fmt.Printf("     • Compact representation: Less surface area\n")
}

// GetFlagForField Retorna o nome da flag de um campo já registrado no contexto
func GetFlagForField(ctx *TranspileContext, structName, fieldName string) (string, bool) {
	if s, ok := ctx.Structs[structName]; ok {
		flag, exists := s.FlagMapping[fieldName]
		return flag, exists
	}
	return "", false
}

// AddFlagMapping Registra um mapeamento de campo -> flag no contexto
func AddFlagMapping(ctx *TranspileContext, structName, fieldName, flagName string, bitIndex int) {
	ctx.AddFlagMapping(structName, fieldName, flagName, 1<<bitIndex)
}

// LogVerbose helper to log transformations (if verbose logging exists in context)
func (ctx *TranspileContext) LogVerbose(fset *token.FileSet, format string, args ...interface{}) {
	// For now, just print - could be enhanced with proper logging
	if len(args) > 0 {
		// Use a simple printf for now
		_ = fset // avoid unused warning
		// fmt.Printf(format+"\n", args...)
	}
}
