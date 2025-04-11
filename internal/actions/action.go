package actions

import "sync"

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
	mu        sync.RWMutex
	ID        string
	Type      string
	Status    string
	Errors    []error
	Results   map[string]interface{}
	isRunning bool
}

// Common methods
func (a *Action) GetID() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.ID
}
func (a *Action) GetType() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Type
}
func (a *Action) GetResults() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Results
}
func (a *Action) GetStatus() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Status
}
func (a *Action) GetErrors() []error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Errors
}
func (a *Action) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.isRunning
}
func (a *Action) CanExecute() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return !a.isRunning
}
