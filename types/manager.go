package types

import (
	l "github.com/faelmori/logz"
	"go/ast"
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
	AddAction(action IAction)
	StartChecking(workerCount int) error
	StopChecking()
	LoadConfig() error
	SaveConfig() error
	CanNotify() bool
	PrepareActions() error
	GetLogger() l.Logger
	GetFilesList(force bool) ([]string, []*ast.File, error)
}
