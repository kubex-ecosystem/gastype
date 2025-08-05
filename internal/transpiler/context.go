// Package transpiler provides context tracking for transpilation operations
package transpiler

import (
	"encoding/json"
	"os"
	"strings"
)

// TranspileContext tracks all information about a transpilation operation
type TranspileContext struct {
	// General configuration
	Ofuscate  bool   `json:"ofuscate"`   // If true, names and structure will be obfuscated
	MapFile   string `json:"map_file"`   // Path to output .map.json file
	InputFile string `json:"input_file"` // Input file path
	OutputDir string `json:"output_dir"` // Output directory

	// Analysis results
	Structs map[string]*StructInfo `json:"structs"` // Original struct → detailed info
	Flags   map[string][]string    `json:"flags"`   // Struct → list of generated flags
}

// StructInfo contains detailed information about each detected struct
type StructInfo struct {
	OriginalName string            `json:"original_name"` // Original name (e.g., Config)
	NewName      string            `json:"new_name"`      // Final name (e.g., ConfigFlags)
	BoolFields   []string          `json:"bool_fields"`   // Original bool fields
	FlagMapping  map[string]string `json:"flag_mapping"`  // BoolField → FlagName
}

// NewContext creates a new transpilation context
func NewContext(inputFile, outputDir string, ofuscate bool, mapFile string) *TranspileContext {
	return &TranspileContext{
		Ofuscate:  ofuscate,
		MapFile:   mapFile,
		InputFile: inputFile,
		OutputDir: outputDir,
		Structs:   make(map[string]*StructInfo),
		Flags:     make(map[string][]string),
	}
}

// AddStruct registers a struct transformation in the context
func (ctx *TranspileContext) AddStruct(originalName, newName string, boolFields []string) {
	mapping := make(map[string]string)
	for _, f := range boolFields {
		mapping[f] = "Flag" + strings.Title(f)
	}

	ctx.Structs[originalName] = &StructInfo{
		OriginalName: originalName,
		NewName:      newName,
		BoolFields:   boolFields,
		FlagMapping:  mapping,
	}
	ctx.Flags[newName] = boolFields
}

// GetFlagName returns the flag name for a given struct and field
func (ctx *TranspileContext) GetFlagName(structName, fieldName string) string {
	if structInfo, exists := ctx.Structs[structName]; exists {
		if flagName, exists := structInfo.FlagMapping[fieldName]; exists {
			return flagName
		}
	}
	return "Flag" + strings.Title(fieldName) // fallback
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
