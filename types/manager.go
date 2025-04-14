package types

import (
	l "github.com/faelmori/logz"
)

type ITypeManager interface {
	GetNotifierChan() chan interface{}
	GetEmail() string
	GetEmailToken() string
	GetNotify() bool
	GetConfig() IConfig
	GetActions() []IAction
	IsRunning() bool
	SetNotifierChan(notifierChan chan interface{})
	SetEmail(email string)
	SetEmailToken(emailToken string)
	SetNotify(notify bool)
	SetConfig(cfg IConfig)
	SetFiles(files []string)
	SetWorkerManager(workerManager IWorker)
	SetLogger(logger l.Logger)
	SetActions(actions []IAction)
	AddAction(action IAction)
	StartChecking(workerCount int) error
	StopChecking()
	LoadConfig() error
	SaveConfig() error
	CanNotify() bool
	PrepareActions() error
	GetLogger() l.Logger
	GetWorkerManager() IWorker
	GetWorkerPool() IWorkerPool
	GetFilesList(force bool) ([]string, error)
}
