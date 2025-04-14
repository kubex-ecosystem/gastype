package types

type IActionBase interface {
	GetType() string
	GetResults() map[string]IResult
	GetStatus() string
	GetErrors() []error
	IsRunning() bool
	CanExecute() bool
}

type IAction interface {
	IActionBase
	Execute() error
}
