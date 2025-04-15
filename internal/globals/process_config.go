package globals

import (
	t "github.com/faelmori/gastype/types"
	l "github.com/faelmori/logz"
)

type ProcessConfig struct {
	Worker     t.IWorker
	Packages   []string
	Config     t.IConfig
	ChanResult chan t.IResult
	ChanError  chan error
	ChanDone   chan bool
	Logger     l.Logger
}

func NewProcessConfig(worker t.IWorker, packages []string, config t.IConfig, logger l.Logger) *ProcessConfig {
	if logger == nil {
		logger = l.GetLogger("GasType")
	}
	return &ProcessConfig{
		Worker:     worker,
		Packages:   packages,
		Config:     config,
		ChanResult: make(chan t.IResult, 50),
		ChanError:  make(chan error, 50),
		ChanDone:   make(chan bool, 50),
		Logger:     logger,
	}
}

func (p *ProcessConfig) WatchResults() t.IResult {
	if p.ChanResult == nil {
		p.ChanResult = make(chan t.IResult, 50)
	}
	return <-p.ChanResult
}
func (p *ProcessConfig) WatchErrors() error {
	if p.ChanError == nil {
		p.ChanError = make(chan error, 50)
	}
	return <-p.ChanError
}
func (p *ProcessConfig) WatchDone() bool {
	if p.ChanDone == nil {
		p.ChanDone = make(chan bool, 50)
	}
	return <-p.ChanDone
}

func (p *ProcessConfig) GetWorker() t.IWorker {
	if p.Worker == nil {
		p.Logger.ErrorCtx("Worker is nil", nil)
		return nil
	}
	return p.Worker
}
func (p *ProcessConfig) GetPackages() []string {
	if p.Packages == nil {
		p.Logger.ErrorCtx("Packages is nil", nil)
		return nil
	}
	return p.Packages
}
func (p *ProcessConfig) GetConfig() t.IConfig {
	if p.Config == nil {
		p.Logger.ErrorCtx("Config is nil", nil)
		return nil
	}
	return p.Config
}
func (p *ProcessConfig) GetChanResult() chan t.IResult {
	if p.ChanResult == nil {
		p.ChanResult = make(chan t.IResult, 50)
	}
	return p.ChanResult
}
func (p *ProcessConfig) GetChanError() chan error {
	if p.ChanError == nil {
		p.ChanError = make(chan error, 50)
	}
	return p.ChanError
}
func (p *ProcessConfig) GetChanDone() chan bool {
	if p.ChanDone == nil {
		p.ChanDone = make(chan bool, 50)
	}
	return p.ChanDone
}
func (p *ProcessConfig) GetLogger() l.Logger {
	if p.Logger == nil {
		p.Logger = l.GetLogger("GasType")
	}
	return p.Logger
}

func (p *ProcessConfig) SetWorker(worker t.IWorker) {
	if worker == nil {
		p.Logger.ErrorCtx("Worker is nil", nil)
		return
	}
	p.Worker = worker
}
func (p *ProcessConfig) SetPackages(packages []string) {
	if packages == nil {
		p.Logger.ErrorCtx("Packages is nil", nil)
		return
	}
	p.Packages = packages
}
func (p *ProcessConfig) SetConfig(config t.IConfig) {
	if config == nil {
		p.Logger.ErrorCtx("Config is nil", nil)
		return
	}
	p.Config = config
}
func (p *ProcessConfig) SetChanResult(chanResult chan t.IResult) {
	if chanResult == nil {
		p.Logger.ErrorCtx("ChanResult is nil", nil)
		return
	}
	p.ChanResult = chanResult
}
func (p *ProcessConfig) SetChanError(chanError chan error) {
	if chanError == nil {
		p.Logger.ErrorCtx("ChanError is nil", nil)
		return
	}
	p.ChanError = chanError
}
func (p *ProcessConfig) SetChanDone(chanDone chan bool) {
	if chanDone == nil {
		p.Logger.ErrorCtx("ChanDone is nil", nil)
		return
	}
	p.ChanDone = chanDone
}
func (p *ProcessConfig) SetLogger(logger l.Logger) {
	if logger == nil {
		p.Logger = l.GetLogger("GasType")
		return
	}
	p.Logger = logger
}
