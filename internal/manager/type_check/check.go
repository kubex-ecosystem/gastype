package type_check

import (
	t "github.com/faelmori/gastype/types"
	l "github.com/faelmori/logz"
)

type CheckProcess struct {
	Worker     t.IWorker
	Packages   []string
	Config     t.IConfig
	ChanError  chan error
	ChanDone   chan bool
	ChanResult chan TypeCheckDetails
	Logger     l.Logger
}

func NewCheckProcess(worker t.IWorker, packages []string, config t.IConfig, logger l.Logger) *CheckProcess {
	if logger == nil {
		logger = l.GetLogger("GasType")
	}
	return &CheckProcess{
		Worker:     worker,
		Packages:   packages,
		Config:     config,
		ChanResult: make(chan TypeCheckDetails, 50),
		ChanError:  make(chan error, 50),
		ChanDone:   make(chan bool, 50),
		Logger:     logger,
	}
}

func (p *CheckProcess) WatchResults() TypeCheckDetails {
	if p.ChanResult == nil {
		p.ChanResult = make(chan TypeCheckDetails, 50)
	}
	return <-p.ChanResult
}
func (p *CheckProcess) WatchErrors() error {
	if p.ChanError == nil {
		p.ChanError = make(chan error, 50)
	}
	return <-p.ChanError
}
func (p *CheckProcess) WatchDone() bool {
	if p.ChanDone == nil {
		p.ChanDone = make(chan bool, 50)
	}
	return <-p.ChanDone
}

func (p *CheckProcess) GetWorker() t.IWorker {
	if p.Worker == nil {
		p.Logger.ErrorCtx("Worker is nil", nil)
		return nil
	}
	return p.Worker
}
func (p *CheckProcess) GetPackages() []string {
	if p.Packages == nil {
		p.Logger.ErrorCtx("Packages is nil", nil)
		return nil
	}
	return p.Packages
}
func (p *CheckProcess) GetConfig() t.IConfig {
	if p.Config == nil {
		p.Logger.ErrorCtx("Config is nil", nil)
		return nil
	}
	return p.Config
}
func (p *CheckProcess) GetChanResult() chan TypeCheckDetails {
	if p.ChanResult == nil {
		p.ChanResult = make(chan TypeCheckDetails, 50)
	}
	return p.ChanResult
}
func (p *CheckProcess) GetChanError() chan error {
	if p.ChanError == nil {
		p.ChanError = make(chan error, 50)
	}
	return p.ChanError
}
func (p *CheckProcess) GetChanDone() chan bool {
	if p.ChanDone == nil {
		p.ChanDone = make(chan bool, 50)
	}
	return p.ChanDone
}
func (p *CheckProcess) GetLogger() l.Logger {
	if p.Logger == nil {
		p.Logger = l.GetLogger("GasType")
	}
	return p.Logger
}

func (p *CheckProcess) SetWorker(worker t.IWorker) {
	if worker == nil {
		p.Logger.ErrorCtx("Worker is nil", nil)
		return
	}
	p.Worker = worker
}
func (p *CheckProcess) SetPackages(packages []string) {
	if packages == nil {
		p.Logger.ErrorCtx("Packages is nil", nil)
		return
	}
	p.Packages = packages
}
func (p *CheckProcess) SetConfig(config t.IConfig) {
	if config == nil {
		p.Logger.ErrorCtx("Config is nil", nil)
		return
	}
	p.Config = config
}
func (p *CheckProcess) SetChanResult(chanResult chan TypeCheckDetails) {
	if chanResult == nil {
		p.Logger.ErrorCtx("ChanResult is nil", nil)
		return
	}
	p.ChanResult = chanResult
}
func (p *CheckProcess) SetChanError(chanError chan error) {
	if chanError == nil {
		p.Logger.ErrorCtx("ChanError is nil", nil)
		return
	}
	p.ChanError = chanError
}
func (p *CheckProcess) SetChanDone(chanDone chan bool) {
	if chanDone == nil {
		p.Logger.ErrorCtx("ChanDone is nil", nil)
		return
	}
	p.ChanDone = chanDone
}
func (p *CheckProcess) SetLogger(logger l.Logger) {
	if logger == nil {
		p.Logger = l.GetLogger("GasType")
		return
	}
	p.Logger = logger
}
