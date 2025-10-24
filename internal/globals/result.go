// Package globals provides a Result struct that encapsulates the result of a type check operation.
package globals

import (
	"fmt"

	t "github.com/kubex-ecosystem/gastype/interfaces"
	gl "github.com/kubex-ecosystem/logz/logger"
)

type Result struct {
	Package string `json:"package"`         // Name of the package
	Status  string `json:"status"`          // Status of the type check (Success, Failed, Error)
	Error   string `json:"error,omitempty"` // Error message if any
}

func NewResult(pkg, status string, err error) t.IResult {
	errorStr := ""

	if err != nil {
		gl.Log("error", fmt.Sprintf("[ %s ] %s", pkg, err.Error()))
		errorStr = err.Error()
	}

	return &Result{
		Package: pkg,
		Status:  status,
		Error:   errorStr,
	}
}

func (c *Result) GetPackage() string { return c.Package }
func (c *Result) GetStatus() string  { return c.Status }
func (c *Result) GetError() string   { return c.Error }

func (c *Result) SetPackage(packageName string) { c.Package = packageName }
func (c *Result) SetStatus(status string)       { c.Status = status }
func (c *Result) SetError(err string)           { c.Error = err }

func (c *Result) ToJSON(outputTarget string) string {
	return ""
}
func (c *Result) ToXML(outputTarget string) string {
	return ""
}
func (c *Result) ToCSV(outputTarget string) string {
	return ""
}
func (c *Result) ToMap() map[string]interface{} {
	return nil
}

func (c *Result) DataTable() error {
	return nil
}
