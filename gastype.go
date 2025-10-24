// Package gastype provides functionalities for type checking Go code and saving results to JSON files.
package gastype

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	t "github.com/kubex-ecosystem/gastype/interfaces"

	gl "github.com/kubex-ecosystem/logz/logger"
)

var ()

// runTypeCheck performs the type checking process.

// performTypeCheck performs type checking for a given package.

// saveResultsToJSON saves the type check results to a JSON file.
func saveResultsToJSON(results []t.IResult, filename string) {
	// Validate and sanitize the output file path
	absFile, err := filepath.Abs(filename)
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Invalid output file path: %v", err))
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error generating JSON: %v", err))
	}

	err = os.WriteFile(absFile, data, 0644)
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error saving JSON: %v", err))
	}

	gl.Log("info", fmt.Sprintf("Results saved to: %s", absFile))
}
