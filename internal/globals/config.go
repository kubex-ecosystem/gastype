package globals

import (
	t "github.com/faelmori/gastype/types"
	s "github.com/faelmori/gkbxsrv/services"
)

type Config struct {
	cfgSrv      s.ConfigService
	dir         string
	workerCount int
	outputFile  string
}

func NewConfig() t.IConfig {
	return &Config{}
}

func (c *Config) Load() error {
	c.dir = "./example"
	c.workerCount = 4
	c.outputFile = "type_check_results.json"
	return nil
}

func (c *Config) GetDir() string        { return c.dir }
func (c *Config) GetWorkerCount() int   { return c.workerCount }
func (c *Config) GetOutputFile() string { return c.outputFile }

func (c *Config) SetDir(dir string)               { c.dir = dir }
func (c *Config) SetWorkerLimit(workerCount int)  { c.workerCount = workerCount }
func (c *Config) SetOutputFile(outputFile string) { c.outputFile = outputFile }
