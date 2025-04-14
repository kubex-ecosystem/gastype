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

func ParseAstFile(fSet *token.FileSet, filePath string, lgr l.Logger) (*ast.File, error) {
	astFile, err := parser.ParseFile(fSet, filePath, nil, parser.ParseComments)
	if err != nil {
		lgr.ErrorCtx(fmt.Sprintf("error parsing file %s: %v", filePath, err), nil)
		return nil, err
	}
	lgr.DebugCtx(fmt.Sprintf("Parsed file: %s", filePath), nil)
	return astFile, nil
}

// CollectGoFiles collects all Go files in a directory recursively
func CollectGoFiles(dirPath string, files *[]string, lgr l.Logger) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		lgr.ErrorCtx(fmt.Sprintf("error reading directory %s: %v", dirPath, err), nil)
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() && entry.Name() != "." && entry.Name() != ".." && entry.Name() != ".git" {
			lgr.DebugCtx(fmt.Sprintf("Recursing into directory: %s", fullPath), nil)
			if err := CollectGoFiles(fullPath, files, lgr); err != nil {
				lgr.ErrorCtx(fmt.Sprintf("error reading subdirectory %s: %v", fullPath, err), nil)
				return err
			}
		} else if filepath.Ext(entry.Name()) == ".go" {
			lgr.DebugCtx(fmt.Sprintf("Found Go file: %s", fullPath), nil)
			*files = append(*files, fullPath)
		}
	}

	return nil
}
