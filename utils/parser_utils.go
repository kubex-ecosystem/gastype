package utils

import (
	"fmt"
	l "github.com/faelmori/logz"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

func collectGoFiles(dirPath string, files *[]string, lgr l.Logger) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		lgr.Error(fmt.Sprintf("error reading directory %s: %v", dirPath, err), nil)
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			// Recursivamente percorre subpastas
			err := collectGoFiles(fullPath, files, lgr)
			if err != nil {
				lgr.Error(fmt.Sprintf("error reading subdirectory %s: %v", fullPath, err), nil)
				return err
			}
		} else if filepath.Ext(entry.Name()) == ".go" {
			lgr.Info(fmt.Sprintf("Found Go file: %s", fullPath), nil)
			*files = append(*files, fullPath)
		}
	}

	return nil
}

// ParseFiles processes all Go files in a directory and returns parsed packages
func ParseFiles(dir string) (map[string][]*ast.File, error) {
	l.Debug(fmt.Sprintf("Parsing Go files in directory: %s", dir), nil)

	filesSet := token.NewFileSet()

	packages := make(map[string][]*ast.File)

	absDir, absDirErr := filepath.Abs(dir)
	l.Debug(fmt.Sprintf("Absolute directory path: %s", absDir), nil)
	if absDirErr != nil {
		l.Error(fmt.Sprintf("Error getting absolute directory path: %s", absDirErr.Error()), nil)
		return nil, fmt.Errorf("invalid directory path: %v", absDirErr)
	}
	// Check if the directory exists
	if _, statErr := os.Stat(absDir); os.IsNotExist(statErr) {
		l.Error(fmt.Sprintf("Directory does not exist: %s", absDir), nil)
		return nil, fmt.Errorf("directory does not exist: %s", absDir)
	}

	// Find Go files
	// Use filepath.Glob to find all Go files in the directory
	// This will match all files with .go extension in the directory
	// and its subdirectories
	// Note: This is a simple implementation, you may want to use a more robust method
	files := make([]string, 0)
	filesErr := collectGoFiles(absDir, &files, l.GetLogger("GasType"))
	if filesErr != nil {
		fmt.Println("Erro ao coletar arquivos:", filesErr)
		return nil, filesErr
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
