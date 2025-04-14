package types

import (
	l "github.com/faelmori/logz"
)

type ITypeManager interface {
	GetConfig() IConfig

	GetNotifierChan() chan interface{}
	GetEmail() string
	GetEmailToken() string
	GetNotify() bool

	SetNotifierChan(chan interface{})
	SetEmail(string)
	SetEmailToken(string)
	SetNotify(bool)
	CanNotify() bool

	IsRunning() bool

	SetConfig(IConfig)
	SetFiles([]string)
	SetLogger(l.Logger)

	GetActions() []IAction
	AddAction(IAction)
	PrepareActions() error

	StartChecking(int) error
	StopChecking()

	LoadConfig() error
	SaveConfig() error

	GetLogger() l.Logger

	GetWorkerManager() IWorker
	GetWorkerPool() IWorkerPool
	SetWorkerManager(IWorker)
	SetWorkerPool(IWorkerPool)

	GetFilesList(bool) ([]string, error)
}
