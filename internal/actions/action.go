package actions

import (
	t "github.com/faelmori/gastype/types"
	"sync"
)

// Base interface for all actions

type IAction interface {
	GetType() string
	Execute() error
	Cancel() error
	CanExecute() bool
	IsRunning() bool
	GetID() string
	GetResults() map[string]t.IResult
	GetStatus() string
	GetErrors() []error
}

// Common struct for all actions

type Action struct {
	mu        sync.RWMutex
	ID        string
	Type      string
	Status    string
	Errors    []error
	Results   map[string]t.IResult
	isRunning bool
}

// NewAction creates a new action
func NewAction(actionType string) IAction {
	return &Action{
		Type:    actionType,
		Status:  "Pending",
		Results: make(map[string]t.IResult),
	}
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
func (a *Action) GetResults() map[string]t.IResult {
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
func (a *Action) Execute() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.isRunning {
		return nil
	}
	a.isRunning = true
	return nil
}
func (a *Action) Cancel() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.isRunning {
		return nil
	}
	a.isRunning = false
	return nil
}
