package globals

import (
	"fmt"
	t "github.com/faelmori/gastype/types"
	l "github.com/faelmori/logz"
	"os"
)

type Result struct {
	Package string `json:"package"`         // Name of the package
	Status  string `json:"status"`          // Status of the type check (Success, Failed, Error)
	Error   string `json:"error,omitempty"` // Error message if any
}

func NewResult(pkg, status string, err error) t.IResult {
	errorStr := ""
	l.GetLogger("GasType").DebugCtx(fmt.Sprintf("[ %s ] %s", pkg, status), map[string]interface{}{"package": pkg, "status": status})
	if err != nil {
		l.GetLogger("GasType").ErrorCtx(fmt.Sprintf("[ %s ] %s", pkg, err.Error()), map[string]interface{}{})
		errorStr = err.Error()
	}
	if status == "" {
		l.GetLogger("GasType").ErrorCtx(fmt.Sprintf("[ %s ] %s", pkg, "Status is empty"), nil)
		status = "Error"
	}
	if pkg == "" {
		l.GetLogger("GasType").ErrorCtx(fmt.Sprintf("[ %s ] %s", pkg, "Package is empty"), nil)
		pkg = "Unknown"
	}
	if errorStr == "" {
		l.GetLogger("GasType").ErrorCtx(fmt.Sprintf("[ %s ] %s", pkg, "Error is empty"), nil)
		errorStr = "No error"
	}
	if status == "Error" {
		l.GetLogger("GasType").ErrorCtx(fmt.Sprintf("[ %s ] %s", pkg, "Status is Error"), nil)
		errorStr = "Error"
	}
	l.GetLogger("GasType").DebugCtx(fmt.Sprintf("[ %s ] %s", pkg, status), map[string]interface{}{"package": pkg, "status": status})
	return &Result{
		Package: pkg,
		Status:  status,
		Error:   errorStr,
	}
}

func (c *Result) GetPackage() string {
	l.GetLogger("GasType").InfoCtx(fmt.Sprintf("Getting package %s", c.Package), nil)
	return c.Package
}
func (c *Result) GetStatus() string {
	l.GetLogger("GasType").InfoCtx(fmt.Sprintf("Getting status %s", c.Status), nil)
	return c.Status
}
func (c *Result) GetError() string {
	l.GetLogger("GasType").InfoCtx(fmt.Sprintf("Getting error %s", c.Error), nil)
	return c.Error
}

func (c *Result) SetPackage(packageName string) {
	l.GetLogger("GasType").InfoCtx(fmt.Sprintf("Setting package %s", packageName), nil)
	c.Package = packageName
}
func (c *Result) SetStatus(status string) {
	l.GetLogger("GasType").InfoCtx(fmt.Sprintf("Setting status %s", status), nil)
	c.Status = status
}
func (c *Result) SetError(err string) {
	l.GetLogger("GasType").InfoCtx(fmt.Sprintf("Setting error %s", err), nil)
	c.Error = err
}

func (c *Result) ToJSON(outputTarget string) string {
	l.GetLogger("GasType").InfoCtx(fmt.Sprintf("Converting to JSON %s", c.Package), nil)
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
	l.GetLogger("GasType").InfoCtx(fmt.Sprintf("Converting to XML %s", c.Package), nil)
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
	l.GetLogger("GasType").InfoCtx(fmt.Sprintf("Converting to CSV %s", c.Package), nil)
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
	l.GetLogger("GasType").InfoCtx(fmt.Sprintf("Converting to map %s", c.Package), nil)
	return map[string]interface{}{
		"package": c.Package,
		"status":  c.Status,
		"error":   c.Error,
	}
}

func (c *Result) DataTable() error {
	l.GetLogger("GasType").InfoCtx(fmt.Sprintf("Converting to DataTable %s", c.Package), nil)
	return nil
}
