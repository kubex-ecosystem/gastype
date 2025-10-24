package actions

import (
	i "github.com/kubex-ecosystem/gastype/interfaces"
)

// Action Common struct for all actions
type Action struct {
	ID        string
	Type      string
	Status    string
	Errors    []error
	Results   map[string]interface{}
	isRunning bool
}

func NewAction(id, actionType string) i.IAction {
	return &Action{
		ID:      id,
		Type:    actionType,
		Status:  "Pending",
		Results: make(map[string]interface{}),
	}
}

func (a *Action) GetID() string                      { return a.ID }
func (a *Action) GetType() string                    { return a.Type }
func (a *Action) GetResults() map[string]interface{} { return a.Results }
func (a *Action) GetStatus() string                  { return a.Status }
func (a *Action) GetErrors() []error                 { return a.Errors }
func (a *Action) IsRunning() bool                    { return a.isRunning }
func (a *Action) CanExecute() bool                   { return !a.isRunning }
func (a *Action) Execute() error {
	a.isRunning = true
	defer func() { a.isRunning = false }()

	// Placeholder for actual execution logic
	return nil
}
