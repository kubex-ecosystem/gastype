package types

import (
	"time"
)

type IActionManager interface {
	GetResults() map[string]IResult
	GetErrorChannel() chan error
	GetDoneChannel() chan struct{}
	GetCancelChannel() chan struct{}
	GetResultsChannel() chan IResult
}

type IActionBase interface {
	GetID() string
	GetType() string

	GetStatus() string
	GetErrors() []error
}

type IAction interface {
	IActionBase
	IActionManager
	IsRunning() bool
	CanExecute() bool
	Execute() error
	Cancel() error
}

type IJob interface {
	IAction
	GetAction() IAction
	GetResults() map[string]IResult
	GetErrorChannel() chan error
	GetDoneChannel() chan struct{}
	GetCancelChannel() chan struct{}
	GetResultsChannel() chan IResult
	GetCreateTime() time.Time
	GetFinishTime() time.Time
	GetCancelTime() time.Time
	SetFinishTime(t time.Time)
	SetCancelTime(t time.Time)
}
