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

	notifierChan = make(chan string)
	email        = "gastype@gmail.com"
	emailToken   = "123456"
	notify       = false
	config       = &Config{
		Dir:         "./example",
		WorkerCount: 4,
		OutputFile:  "type_check_results.json",
	}
)

// CheckResult stores the results of the type checking process.
type CheckResult struct {
	Package string `json:"package"`         // Name of the package
	Status  string `json:"status"`          // Status of the type check (Success, Failed, Error)
	Error   string `json:"error,omitempty"` // Error message if any
}

type Config struct {
	Dir         string
	WorkerCount int
	OutputFile  string
}

// main is the entry point of the program.
func main() {
	// Define the root command
	var rootCmd = &cobra.Command{
		Use:     "gastype",
		Short:   "Parallel type checking for Go files",
		Long:    "Parallel type checking for Go files in a given directory",
		Example: `gastype -d ./example -w 4 -o type_check_results.json`,
		Aliases: []string{"got", "gotype"},
		Run:     runTypeCheck,
	}

	var checkCmd = &cobra.Command{
		Use:     "check",
		Short:   "Check Go files for type errors",
		Long:    "Check Go files for type errors in a given directory",
		Example: `gastype check -d ./example -w 4 -o type_check_results.json`,
		Run:     runTypeCheck,
	}

	// Add flags to the root command
	checkCmd.Flags().StringVarP(&dir, "dir", "d", "./example", "Directory containing Go files")
	checkCmd.Flags().IntVarP(&workerCount, "workers", "w", 4, "Number of workers for parallel processing")
	checkCmd.Flags().StringVarP(&outputFile, "output", "o", "type_check_results.json", "Output file for JSON results")

	// Add commands to the root command
	rootCmd.AddCommand(checkCmd)

	// Define the watch command
	var watch = &cobra.Command{
		Use:     "watch",
		Short:   "Watcher and notifier for type checking Go files",
		Long:    "Watcher and notifier for type checking Go files in a given directory",
		Example: `gastype watch -d ./example -w 4 -o type_check_results.json`,
		Run:     runTypeCheck,
	}

	// Add flags to the watch command
	watch.Flags().StringVarP(&email, "email", "e", "gastype@gmail.com", "Email address for notifications")
	watch.Flags().StringVarP(&emailToken, "token", "t", "123456", "Token for email notifications")
	watch.Flags().BoolVarP(&notify, "notify", "n", false, "Enable email notifications")

	// Add commands to the root command
	rootCmd.AddCommand(watch)

	// Define command-line flags
	rootCmd.Flags().StringVarP(&dir, "dir", "d", "./example", "Directory containing Go files")
	rootCmd.Flags().IntVarP(&workerCount, "workers", "w", 4, "Number of workers for parallel processing")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "type_check_results.json", "Output file for JSON results")

	SetUsageDefinition(rootCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// runTypeCheck performs the type checking process.
func runTypeCheck(_ *cobra.Command, _ []string) {
	// Validate and sanitize the directory path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("Invalid directory path: %v", err)
	}

	// Ensure the directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", absDir)
	}

	// Find all Go files in the directory
	files, err := filepath.Glob(filepath.Join(absDir, "*.go"))
	if err != nil {
		log.Fatal(err)
	}

	outputFileSanitized := filepath.Clean(outputFile)
	if outputFileSanitized == "" {
		log.Fatal("Invalid output file path")
	}

	filesSet := token.NewFileSet()
	packages := make(map[string][]*ast.File)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Channel to capture parsing errors
	errChan := make(chan error, len(files))

	// Parse files in parallel
	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			// Validate and sanitize the file path
			absFile, err := filepath.Abs(file)
			if err != nil {
				errChan <- fmt.Errorf("invalid file path: %v", err)
				return
			}

			src, err := os.ReadFile(absFile)
			if err != nil {
				errChan <- fmt.Errorf("error reading file %s: %v", absFile, err)
				return
			}

			node, err := parser.ParseFile(filesSet, absFile, src, parser.AllErrors)
			if err != nil {
				errChan <- fmt.Errorf("error parsing %s: %v", absFile, err)
				return
			}

			// Store files grouped by package
			mu.Lock()
			packages[node.Name.Name] = append(packages[node.Name.Name], node)
			mu.Unlock()
		}(file)
	}

	wg.Wait()
	close(errChan)

	// Display parsing errors
	for err := range errChan {
		log.Println(err)
	}

	// If there are no packages, abort
	if len(packages) == 0 {
		log.Fatal("No files parsed successfully.")
	}

	// Sort packages by size (priority)
	sortedPackages := make([]string, 0, len(packages))
	for pkgName := range packages {
		sortedPackages = append(sortedPackages, pkgName)
	}
	sort.Slice(sortedPackages, func(i, j int) bool {
		return len(packages[sortedPackages[i]]) > len(packages[sortedPackages[j]])
	})

	// Create workers for type checking
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

	// Send packages for type checking (largest first)
	for _, pkgName := range sortedPackages {
		typeCheckJobs <- pkgName
	}
	close(typeCheckJobs)

	// Wait for workers to finish
	workers.Wait()
	close(results)

	// Collect results
	var checkResults []CheckResult
	for res := range results {
		checkResults = append(checkResults, res)
		fmt.Println(res.Status)
	}

	// Save results to JSON
	saveResultsToJSON(checkResults, outputFileSanitized)
}

// performTypeCheck performs type checking for a given package.
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

// saveResultsToJSON saves the type check results to a JSON file.
func saveResultsToJSON(results []CheckResult, filename string) {
	// Validate and sanitize the output file path
	absFile, err := filepath.Abs(filename)
	if err != nil {
		log.Fatalf("Invalid output file path: %v", err)
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("Error generating JSON: %v", err)
	}

	err = os.WriteFile(absFile, data, 0644)
	if err != nil {
		log.Fatalf("Error saving JSON: %v", err)
	}

	fmt.Println("Results saved to:", absFile)
}
