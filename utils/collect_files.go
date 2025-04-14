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

// CollectGoFiles collects all Go files in a directory recursively
func CollectGoFiles(dirPath string, files *[]string, astFiles *[]*ast.File, lgr l.Logger) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		lgr.ErrorCtx(fmt.Sprintf("error reading directory %s: %v", dirPath, err), nil)
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			lgr.DebugCtx(fmt.Sprintf("Recursing into directory: %s", fullPath), nil)
			if err := CollectGoFiles(fullPath, files, astFiles, lgr); err != nil {
				lgr.ErrorCtx(fmt.Sprintf("error reading subdirectory %s: %v", fullPath, err), nil)
				return err
			}
		} else if filepath.Ext(entry.Name()) == ".go" {
			lgr.DebugCtx(fmt.Sprintf("Found Go file: %s", fullPath), nil)
			*files = append(*files, fullPath)
			fset := token.NewFileSet()
			astFile, err := parser.ParseFile(fset, fullPath, nil, parser.ParseComments)
			if err != nil {
				lgr.ErrorCtx(fmt.Sprintf("error parsing file %s: %v", fullPath, err), nil)
				return err
			}
			*astFiles = append(*astFiles, astFile)
		}
	}

	return nil
}
