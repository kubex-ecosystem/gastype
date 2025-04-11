package globals

import (
	"fmt"
	t "github.com/faelmori/gastype/types"
	s "github.com/faelmori/gkbxsrv/services"
	l "github.com/faelmori/logz"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	cfgSrv      s.ConfigService
	dir         string
	workerCount int
	outputFile  string
	enviroment  t.IEnvironment
}

func NewConfig() t.IConfig {
	cfg := &Config{}
	cfg.enviroment = NewEnvironment()
	return cfg
}

func NewConfigWithArgs(dir string, workerCount int, outputFile string) t.IConfig {
	l.Notice("Creating configuration", nil)
	cfg := NewConfig()
	cfg.SetDir(dir)
	cfg.SetWorkerLimit(workerCount)
	cfg.SetOutputFile(outputFile)

	if loadErr := cfg.Load(); loadErr != nil {
		l.Error("Error loading configuration", map[string]interface{}{"error": loadErr.Error()})
		return nil
	} else {
		l.Success("Configuration loaded successfully", nil)
		return cfg
	}
}

func (c *Config) Load() error {
	if c.dir == "" {
		l.Notice("Loading configuration from environment", nil)
		dir, dirErr := filepath.Abs("./")
		if dirErr != nil {
			l.Error("Error getting absolute path", map[string]interface{}{"error": dirErr.Error()})
			return dirErr
		}
		c.dir = dir
	}
	l.Notice(fmt.Sprintf("Configuration loaded from %s", c.dir), nil)

	l.Notice(fmt.Sprintf("Begining with %d workers", c.workerCount), map[string]interface{}{"workerCount": c.workerCount})
	if c.workerCount == 0 {
		cpuCount := c.enviroment.CpuCount()
		if cpuCount <= 0 {
			cpuCount = runtime.NumCPU()
		}
		if cpuCount > 0 {
			if cpuCount >= 4 {
				c.workerCount = 4
			} else {
				c.workerCount = cpuCount
			}
		}
	}
	l.Notice(fmt.Sprintf("Worker count set to %d", c.workerCount), map[string]interface{}{"workerCount": c.workerCount})

	l.Notice(fmt.Sprintf("Searching for output file: %s", c.outputFile), map[string]interface{}{"outputFile": c.outputFile})
	if c.outputFile == "" {
		homeDir, homeDirErr := os.UserHomeDir()
		if homeDirErr != nil {
			homeDir, homeDirErr = os.UserCacheDir()
			if homeDirErr != nil {
				c.outputFile = "type_check_results.json"
			} else {
				c.outputFile = homeDir + "/tmp/type_check_results.json"
			}
		}
	}
	l.Notice(fmt.Sprintf("Output file set to: %s", c.outputFile), map[string]interface{}{"outputFile": c.outputFile})

	l.Notice("Configuration loaded from environment", nil)
	l.Notice("========================================", nil)
	l.Notice("Directory: "+c.dir, nil)
	l.Notice("Worker Count: "+string(c.workerCount), nil)
	l.Notice("Output File: "+c.outputFile, nil)
	l.Notice("Environment CPU Count: "+string(c.enviroment.CpuCount()), nil)
	l.Notice("Environment OS: "+c.enviroment.Os(), nil)
	l.Notice("Environment Hostname: "+c.enviroment.Hostname(), nil)
	l.Notice("Environment Mem Total: "+string(c.enviroment.MemTotal()), nil)
	l.Notice("Environment Kernel: "+c.enviroment.Kernel(), nil)
	l.Notice("========================================", nil)
	l.Success("Configuration loaded successfully", nil)

	return nil
}
func (c *Config) GetDir() string                  { return c.dir }
func (c *Config) GetWorkerCount() int             { return c.workerCount }
func (c *Config) GetOutputFile() string           { return c.outputFile }
func (c *Config) SetDir(dir string)               { c.dir = dir }
func (c *Config) SetWorkerLimit(workerCount int)  { c.workerCount = workerCount }
func (c *Config) SetOutputFile(outputFile string) { c.outputFile = outputFile }
