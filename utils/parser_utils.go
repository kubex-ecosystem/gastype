package utils

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	gl "github.com/kubex-ecosystem/logz/logger"
)

// ParseFiles processes all Go files in a directory and returns parsed packages
func ParseFiles(dir string) (map[string][]*ast.File, error) {
	filesSet := token.NewFileSet()

	packages := make(map[string][]*ast.File)

	// Ensure directory exists
	absDir, absDirErr := filepath.Abs(dir)
	if absDirErr != nil {
		gl.Log("error", fmt.Sprintf("Invalid directory path: %v", absDirErr))
		return nil, fmt.Errorf("invalid directory path: %v", absDirErr)
	}
	if _, statErr := os.Stat(absDir); os.IsNotExist(statErr) {
		gl.Log("error", fmt.Sprintf("Directory does not exist: %s", absDir))
		return nil, fmt.Errorf("directory does not exist: %s", absDir)
	}

	// Find Go files
	files, globErr := filepath.Glob(filepath.Join(absDir, "*.go"))
	if globErr != nil {
		gl.Log("error", fmt.Sprintf("Error finding Go files: %s", globErr.Error()))
		return nil, fmt.Errorf("error finding Go files: %v", globErr)
	}
	if len(files) == 0 {
		lenFilesErr := fmt.Errorf("no Go files found in directory: %s", absDir)
		gl.Log("error", fmt.Sprintf("Error parsing Go files: %s", lenFilesErr.Error()))
		return nil, lenFilesErr
	}

	// Parse files
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			gl.Log("error", fmt.Sprintf("Error reading file %s: %v", file, err))
			return nil, fmt.Errorf("error reading file %s: %v", file, err)
		}

		node, err := parser.ParseFile(filesSet, file, src, parser.AllErrors)
		if err != nil {
			gl.Log("error", fmt.Sprintf("Error parsing file %s: %v", file, err))
			return nil, fmt.Errorf("error parsing file %s: %v", file, err)
		}

		pkgName := node.Name.Name
		packages[pkgName] = append(packages[pkgName], node)
	}

	return packages, nil
}
