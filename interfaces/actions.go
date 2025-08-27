package interfaces

type IActionBase interface {
	GetType() string
	GetResults() map[string]interface{}
	GetStatus() string
	GetErrors() []error
	IsRunning() bool
	CanExecute() bool
}

type IAction interface {
	IActionBase
	Execute() error
}
