package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/spf13/cobra"
)

var (
	dir         string
	workerCount int
	outputFile  string
)

// CheckResult stores the results of the type checking process.
type CheckResult struct {
	Package string `json:"package"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "typechecker",
		Short: "Parallel type checking for Go files",
		Run:   runTypeCheck,
	}

	rootCmd.Flags().StringVarP(&dir, "dir", "d", "./example", "Directory containing Go files")
	rootCmd.Flags().IntVarP(&workerCount, "workers", "w", 4, "Number of workers for parallel processing")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "type_check_results.json", "Output file for JSON results")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runTypeCheck(cmd *cobra.Command, args []string) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("Invalid directory path: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(absDir, "*.go"))
	if err != nil {
		log.Fatal(err)
	}

	filesSet := token.NewFileSet()
	packages := make(map[string][]*ast.File)
	var mu sync.Mutex
	var wg sync.WaitGroup

	errChan := make(chan error, len(files))

	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			src, err := os.ReadFile(file)
			if err != nil {
				errChan <- fmt.Errorf("error reading file %s: %v", file, err)
				return
			}

			node, err := parser.ParseFile(filesSet, file, src, parser.AllErrors)
			if err != nil {
				errChan <- fmt.Errorf("error parsing %s: %v", file, err)
				return
			}

			mu.Lock()
			packages[node.Name.Name] = append(packages[node.Name.Name], node)
			mu.Unlock()
		}(file)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		log.Println(err)
	}

	if len(packages) == 0 {
		log.Fatal("No files parsed successfully.")
	}

	sortedPackages := make([]string, 0, len(packages))
	for pkgName := range packages {
		sortedPackages = append(sortedPackages, pkgName)
	}
	sort.Slice(sortedPackages, func(i, j int) bool {
		return len(packages[sortedPackages[i]]) > len(packages[sortedPackages[j]])
	})

	typeCheckJobs := make(chan string, len(packages))
	results := make(chan CheckResult, len(packages))

	var workers sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for pkgName := range typeCheckJobs {
				performTypeCheck(pkgName, filesSet, packages[pkgName], results)
			}
		}()
	}

	for _, pkgName := range sortedPackages {
		typeCheckJobs <- pkgName
	}
	close(typeCheckJobs)

	workers.Wait()
	close(results)

	var checkResults []CheckResult
	for res := range results {
		checkResults = append(checkResults, res)
		fmt.Println(res.Status)
	}

	saveResultsToJSON(checkResults, outputFile)
}

func performTypeCheck(pkgName string, filesSet *token.FileSet, files []*ast.File, results chan<- CheckResult) {
	conf := types.Config{
		Error: func(err error) {
			results <- CheckResult{Package: pkgName, Status: "Error ❌", Error: err.Error()}
		},
	}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	_, err := conf.Check(pkgName, filesSet, files, info)
	if err != nil {
		results <- CheckResult{Package: pkgName, Status: "Failed ❌", Error: err.Error()}
	} else {
		results <- CheckResult{Package: pkgName, Status: "Success ✅"}
	}
}

func saveResultsToJSON(results []CheckResult, filename string) {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("Error generating JSON: %v", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		log.Fatalf("Error saving JSON: %v", err)
	}

	fmt.Println("Results saved to:", filename)
}
