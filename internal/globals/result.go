package globals

import (
	"fmt"
	t "github.com/faelmori/gastype/types"
	l "github.com/faelmori/logz"
	"go/ast"
	"os"
)

type Result struct {
	Package string    `json:"package"`            // Name of the package
	Status  string    `json:"status"`             // Status of the type check (Success, Failed, Error)
	Error   string    `json:"error,omitempty"`    // Error message if any
	ASTFile *ast.File `json:"ast_file,omitempty"` // AST file if available
	Info    t.IInfo   `json:"info,omitempty"`     // Additional information
}

func NewResult(pkg, status string, err error, astFile *ast.File) t.IResult {
	errorStr := ""
	l.GetLogger("GasType").DebugCtx(fmt.Sprintf("[ %s ] %s", pkg, status), map[string]interface{}{"package": pkg, "status": status})
	if err != nil {
		l.GetLogger("GasType").ErrorCtx(fmt.Sprintf("[ %s ] %s", pkg, err.Error()), map[string]interface{}{})
		errorStr = err.Error()
		status = "Error"
	}
	if status == "" {
		l.GetLogger("GasType").ErrorCtx(fmt.Sprintf("[ %s ] %s", pkg, "Status is empty"), nil)
		status = "Error"
	}
	if pkg == "" {
		l.GetLogger("GasType").ErrorCtx(fmt.Sprintf("[ %s ] %s", pkg, "Package is empty"), nil)
		pkg = "Unknown"
	}
	l.GetLogger("GasType").DebugCtx(fmt.Sprintf("[ %s ] %s", pkg, status), map[string]interface{}{"package": pkg, "status": status})
	return &Result{
		Package: pkg,
		Status:  status,
		Error:   errorStr,
		ASTFile: astFile,
		Info:    NewInfo(astFile),
	}
}

func (c *Result) GetPackage() string { return c.Package }
func (c *Result) GetStatus() string  { return c.Status }
func (c *Result) GetError() string   { return c.Error }
func (c *Result) GetAst() interface{} {
	if c.ASTFile != nil {
		return c.ASTFile
	}
	return nil
}
func (c *Result) GetAstFile() string {
	if c.ASTFile != nil {
		return c.ASTFile.Name.Name
	}
	return ""
}
func (c *Result) GetInfo() t.IInfo { return c.Info }

func (c *Result) SetPackage(packageName string) { c.Package = packageName }
func (c *Result) SetStatus(status string)       { c.Status = status }
func (c *Result) SetError(err string)           { c.Error = err }

func (c *Result) ToJSON(outputTarget string) string {
	l.GetLogger("GasType").DebugCtx(fmt.Sprintf("Converting to JSON %s", c.Package), nil)
	if outputTarget == "" {
		pwd, pwdErr := os.Getwd()
		if pwdErr != nil {
			l.GetLogger("GasType").ErrorCtx(fmt.Sprintf("Error getting current directory: %s", pwdErr.Error()), nil)
			return ""
		}
		outputTarget = fmt.Sprintf("%s/%s.json", pwd, c.Package)
	}
	return fmt.Sprintf("%s/%s.json", outputTarget, c.Package)
}
func (c *Result) ToXML(outputTarget string) string {
	l.GetLogger("GasType").DebugCtx(fmt.Sprintf("Converting to XML %s", c.Package), nil)
	if outputTarget == "" {
		pwd, pwdErr := os.Getwd()
		if pwdErr != nil {
			l.GetLogger("GasType").ErrorCtx(fmt.Sprintf("Error getting current directory: %s", pwdErr.Error()), nil)
			return ""
		}
		outputTarget = fmt.Sprintf("%s/%s.xml", pwd, c.Package)
	}
	return fmt.Sprintf("%s/%s.xml", outputTarget, c.Package)
}
func (c *Result) ToCSV(outputTarget string) string {
	l.GetLogger("GasType").DebugCtx(fmt.Sprintf("Converting to CSV %s", c.Package), nil)
	if outputTarget == "" {
		pwd, pwdErr := os.Getwd()
		if pwdErr != nil {
			l.ErrorCtx(fmt.Sprintf("Error getting current directory: %s", pwdErr.Error()), nil)
			return ""
		}
		outputTarget = fmt.Sprintf("%s/%s.csv", pwd, c.Package)
	}
	return fmt.Sprintf("%s/%s.csv", outputTarget, c.Package)
}
func (c *Result) ToMap() map[string]interface{} {
	l.GetLogger("GasType").DebugCtx(fmt.Sprintf("Converting to map %s", c.Package), nil)
	return map[string]interface{}{
		"package": c.Package,
		"status":  c.Status,
		"error":   c.Error,
	}
}

func (c *Result) DataTable() error {
	l.GetLogger("GasType").DebugCtx(fmt.Sprintf("Converting to DataTable %s", c.Package), nil)
	return nil
}
func (c *Result) GetStatusCode() int {
	l.GetLogger("GasType").DebugCtx(fmt.Sprintf("Getting status code %s", c.Package), nil)
	switch c.Status {
	case "Success":
		return 0
	case "Failed":
		return 1
	case "Error":
		return 2
	default:
		return 3
	}
}
func (c *Result) GetStatusText() string {
	l.GetLogger("GasType").DebugCtx(fmt.Sprintf("Getting status text %s", c.Package), nil)
	switch c.Status {
	case "Success":
		return "Success"
	case "Failed":
		return "Failed"
	case "Error":
		return "Error"
	default:
		return "Unknown"
	}
}
