package gastype

import (
	"encoding/json"
	"fmt"
	t "github.com/faelmori/gastype/types"
	"log"
	"os"
	"path/filepath"
)

var ()

// runTypeCheck performs the type checking process.

// performTypeCheck performs type checking for a given package.

// saveResultsToJSON saves the type check results to a JSON file.
func saveResultsToJSON(results []t.IResult, filename string) {
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
