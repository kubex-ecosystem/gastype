package actions

// Base interface for all actions
type IAction interface {
	GetType() string
	Execute() error
	Cancel() error
	CanExecute() bool
	IsRunning() bool
}

// Common struct for all actions
type Action struct {
	ID        string
	Type      string
	Status    string
	Errors    []error
	Results   map[string]interface{}
	isRunning bool
}

// Common methods
func (a *Action) GetType() string                    { return a.Type }
func (a *Action) GetResults() map[string]interface{} { return a.Results }
func (a *Action) GetStatus() string                  { return a.Status }
func (a *Action) GetErrors() []error                 { return a.Errors }
func (a *Action) IsRunning() bool                    { return a.isRunning }
func (a *Action) CanExecute() bool                   { return !a.isRunning }
