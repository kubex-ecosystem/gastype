package globals

import (
	t "github.com/faelmori/gastype/types"
	s "github.com/faelmori/gkbxsrv/services"
	"os"
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
	env := NewEnvironment()
	cfg := &Config{
		dir:         dir,
		workerCount: workerCount,
		outputFile:  outputFile,
		enviroment:  env,
	}

	if loadErr := cfg.Load(); loadErr != nil {
		return nil
	}

	return cfg
}

func (c *Config) Load() error {
	if c.dir == "" {
		c.dir = "./"
	}
	if c.workerCount == 0 {
		cpuCount := c.enviroment.CpuCount()
		if cpuCount > 0 {
			if cpuCount >= 4 {
				c.workerCount = 4
			} else {
				c.workerCount = cpuCount
			}
		}
	}
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
	return nil
}
func (c *Config) GetDir() string                  { return c.dir }
func (c *Config) GetWorkerCount() int             { return c.workerCount }
func (c *Config) GetOutputFile() string           { return c.outputFile }
func (c *Config) SetDir(dir string)               { c.dir = dir }
func (c *Config) SetWorkerLimit(workerCount int)  { c.workerCount = workerCount }
func (c *Config) SetOutputFile(outputFile string) { c.outputFile = outputFile }
