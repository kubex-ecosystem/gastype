package utils

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

// ParseFiles processes all Go files in a directory and returns parsed packages
func ParseFiles(dir string) (map[string][]*ast.File, error) {
	filesSet := token.NewFileSet()
	packages := make(map[string][]*ast.File)

	// Ensure directory exists
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("invalid directory path: %v", err)
	}
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", absDir)
	}

	// Find Go files
	files, err := filepath.Glob(filepath.Join(absDir, "*.go"))
	if err != nil {
		return nil, fmt.Errorf("error finding Go files: %v", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no Go files found in directory: %s", absDir)
	}

	// Parse files
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %v", file, err)
		}

		node, err := parser.ParseFile(filesSet, file, src, parser.AllErrors)
		if err != nil {
			return nil, fmt.Errorf("error parsing file %s: %v", file, err)
		}

		pkgName := node.Name.Name
		packages[pkgName] = append(packages[pkgName], node)
	}

	return packages, nil
}
