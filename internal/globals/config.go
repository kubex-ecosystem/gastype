package globals

import (
	"fmt"
	t "github.com/faelmori/gastype/types"
	c "github.com/faelmori/kubex-interfaces/config"
	m "github.com/faelmori/kubex-interfaces/module"
	l "github.com/faelmori/logz"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	cfgSrv      c.Manager[m.KubexModule]
	dir         string
	workerCount int
	workerLimit int
	outputFile  string
	results     []t.IResult
	logger      l.Logger
	chanCtl     chan interface{}
	chanResult  chan t.IResult
	environment t.IEnvironment
}

func NewConfig[T m.KubexModule](m T, logger l.Logger) t.IConfig {
	if logger == nil {
		logger = l.GetLogger("GasType")
	}
	cfg := &Config{
		logger:     logger,
		chanCtl:    make(chan interface{}, 50),
		chanResult: make(chan t.IResult, 50),
		results:    make([]t.IResult, 0),
	}
	cfgFilePath := os.Getenv("GASTYPE_CONFIG_FILE")
	if cfgFilePath != "" {
		var err error
		cfgFilePath, err = filepath.Abs(cfgFilePath)
		if err != nil {
			logger.ErrorCtx("Error getting absolute path", map[string]interface{}{"error": err.Error()})
			return nil
		}
		if _, err := os.Stat(cfgFilePath); os.IsNotExist(err) {
			logger.WarnCtx("Configuration file does not exist", map[string]interface{}{"error": err.Error()})
			logger.NoticeCtx("Creating default configuration file", nil)
			if err := os.MkdirAll(filepath.Dir(cfgFilePath), os.ModePerm); err != nil {
				logger.ErrorCtx("Error creating configuration directory", map[string]interface{}{"error": err.Error()})
				return nil
			}
			if _, err := os.Create(cfgFilePath); err != nil {
				logger.ErrorCtx("Error creating configuration file", map[string]interface{}{"error": err.Error()})
				return nil
			}
			logger.NoticeCtx(fmt.Sprintf("Configuration file created at %s", cfgFilePath), nil)
		} else {
			logger.NoticeCtx(fmt.Sprintf("Configuration file found at %s", cfgFilePath), nil)
			if err := os.Chmod(cfgFilePath, 0644); err != nil {
				logger.ErrorCtx("Error setting permissions for configuration file", map[string]interface{}{"error": err.Error()})
				return nil
			}
			logger.NoticeCtx(fmt.Sprintf("Permissions set for configuration file at %s", cfgFilePath), nil)
		}
		cfg.dir = filepath.Dir(cfgFilePath)
		logger.NoticeCtx(fmt.Sprintf("Loading configuration from %s", cfgFilePath), nil)
	}
	cfg.environment = NewEnvironment()
	cfg.cfgSrv = c.NewConfigManager[T](m)
	if setLoggerErr := cfg.cfgSrv.SetLogger(logger); setLoggerErr != nil {
		logger.ErrorCtx("Error setting up logger", map[string]interface{}{"error": setLoggerErr.Error()})
		return nil
	}
	logger.NoticeCtx(fmt.Sprintf("Config path: %s", cfg.cfgSrv), nil)
	return cfg
}

func NewConfigWithArgs[T m.KubexModule](dir string, workerLimit int, outputFile string, logger l.Logger, m T) t.IConfig {
	if logger == nil {
		logger = l.GetLogger("GasType")
	}
	logger.NoticeCtx("Creating configuration", nil)
	cfg := NewConfig[T](m, l.GetLogger("GasType"))
	cfg.SetDir(dir)
	cfg.SetWorkerLimit(workerLimit)
	cfg.SetOutputFile(outputFile)

	if loadErr := cfg.Load(); loadErr != nil {
		logger.ErrorCtx("Error loading configuration", map[string]interface{}{"error": loadErr.Error()})
		return nil
	} else {
		logger.SuccessCtx("Configuration loaded successfully", nil)
		return cfg
	}
}

func (c *Config) Load() error {
	if c.dir == "" {
		c.logger.NoticeCtx("Loading configuration from environment", nil)
		dir, dirErr := filepath.Abs("./")
		if dirErr != nil {
			c.logger.ErrorCtx("Error getting absolute path", map[string]interface{}{"error": dirErr.Error()})
			return dirErr
		}
		c.dir = dir
	}
	c.logger.NoticeCtx(fmt.Sprintf("Configuration loaded from %s", c.dir), nil)

	c.logger.NoticeCtx(fmt.Sprintf("Begining with %d workers", c.workerCount), map[string]interface{}{"workerCount": c.workerCount})
	if c.workerCount == 0 {
		cpuCount := c.environment.CpuCount()
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
	c.logger.NoticeCtx(fmt.Sprintf("Worker count set to %d", c.workerCount), map[string]interface{}{"workerCount": c.workerCount})

	c.logger.NoticeCtx(fmt.Sprintf("Searching for output file: %s", c.outputFile), map[string]interface{}{"outputFile": c.outputFile})
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
	c.logger.NoticeCtx(fmt.Sprintf("Output file set to: %s", c.outputFile), map[string]interface{}{"outputFile": c.outputFile})

	if c.results == nil {
		c.results = make([]t.IResult, 0)
	}

	if c.chanResult == nil {
		c.chanResult = make(chan t.IResult, 50)
	}

	c.logger.NoticeCtx("Configuration loaded from environment", nil)
	c.logger.NoticeCtx("========================================", nil)
	c.logger.NoticeCtx("Directory: "+c.dir, nil)
	c.logger.NoticeCtx("Worker Count: "+fmt.Sprintf("%d", c.workerCount), nil)
	c.logger.NoticeCtx("Output File: "+c.outputFile, nil)
	c.logger.NoticeCtx("Environment CPU Count: "+fmt.Sprintf("%d", c.environment.CpuCount()), nil)
	c.logger.NoticeCtx("Environment OS: "+c.environment.Os(), nil)
	c.logger.NoticeCtx("Environment Hostname: "+c.environment.Hostname(), nil)
	c.logger.NoticeCtx("Environment Mem Total: "+fmt.Sprintf("%d", c.environment.MemTotal()), nil)
	c.logger.NoticeCtx("Environment Kernel: "+c.environment.Kernel(), nil)
	c.logger.NoticeCtx("========================================", nil)

	return nil
}
func (c *Config) GetDir() string        { return c.dir }
func (c *Config) GetWorkerCount() int   { return c.workerCount }
func (c *Config) GetWorkerLimit() int   { return c.workerLimit }
func (c *Config) GetOutputFile() string { return c.outputFile }
func (c *Config) SetDir(dir string)     { c.dir = dir }
func (c *Config) SetWorkerLimit(workerLimit int) {
	if workerLimit <= 0 {
		workerLimit = 1
	}
	c.workerCount = workerLimit
	c.workerLimit = workerLimit
}
func (c *Config) SetOutputFile(outputFile string) { c.outputFile = outputFile }
