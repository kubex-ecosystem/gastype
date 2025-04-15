package actions

import (
	t "github.com/faelmori/gastype/types"
	"sync"
)

// IAction defines the interface for an action with methods for execution, cancellation, and status retrieval.
type IAction interface {
	// GetID retrieves the unique identifier of the action.
	GetID() string

	// GetType retrieves the type of the action.
	GetType() string

	// Execute performs the action and returns an error if it fails.
	Execute() error

	// Cancel cancels the action and returns an error if it fails.
	Cancel() error

	// CanExecute checks if the action can be executed.
	CanExecute() bool

	// IsRunning checks if the action is currently running.
	IsRunning() bool

	// GetResults retrieves a map of results associated with the action.
	GetResults() map[string]t.IResult

	// GetStatus retrieves the current status of the action.
	GetStatus() string

	// GetErrors retrieves a list of errors associated with the action.
	GetErrors() []error
}

// Action is a common struct that implements the IAction interface.
type Action struct {
	mu        sync.RWMutex         // Mutex to ensure thread-safe access to the struct fields.
	ID        string               // Unique identifier of the action.
	Type      string               // Type of the action.
	Status    string               // Current status of the action.
	Errors    []error              // List of errors associated with the action.
	Results   map[string]t.IResult // Map of results associated with the action.
	isRunning bool                 // Indicates whether the action is currently running.
}

// NewAction creates a new action with the specified type.
// Parameters:
//   - actionType: The type of the action to create.
//
// Returns:
//   - IAction: A new instance of the Action struct.
func NewAction(actionType string) IAction {
	return &Action{
		Type:    actionType,
		Status:  "Pending",
		Results: make(map[string]t.IResult),
	}
}

// GetID retrieves the unique identifier of the action.
// Returns:
//   - string: The unique identifier of the action.
func (a *Action) GetID() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.ID
}

// GetType retrieves the type of the action.
// Returns:
//   - string: The type of the action.
func (a *Action) GetType() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Type
}

// GetResults retrieves a map of results associated with the action.
// Returns:
//   - map[string]t.IResult: The map of results.
func (a *Action) GetResults() map[string]t.IResult {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Results
}

// GetStatus retrieves the current status of the action.
// Returns:
//   - string: The current status of the action.
func (a *Action) GetStatus() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Status
}

// GetErrors retrieves a list of errors associated with the action.
// Returns:
//   - []error: The list of errors.
func (a *Action) GetErrors() []error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Errors
}

// IsRunning checks if the action is currently running.
// Returns:
//   - bool: True if the action is running, false otherwise.
func (a *Action) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.isRunning
}

// CanExecute checks if the action can be executed.
// Returns:
//   - bool: True if the action can be executed, false otherwise.
func (a *Action) CanExecute() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return !a.isRunning
}

// Execute performs the action and sets its running state to true.
// Returns:
//   - error: An error if the action is already running, nil otherwise.
func (a *Action) Execute() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.isRunning {
		return nil
	}
	a.isRunning = true
	return nil
}

// Cancel cancels the action and sets its running state to false.
// Returns:
//   - error: An error if the action is not running, nil otherwise.
func (a *Action) Cancel() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.isRunning {
		return nil
	}
	a.isRunning = false
	return nil
}

// GetResultChannel retrieves the channel used to communicate action results.
// Returns:
//   - chan t.IResult: The result channel (currently nil).
func (a *Action) GetResultChannel() chan t.IResult {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return nil
}

// GetErrorChannel retrieves the channel used to communicate errors.
// Returns:
//   - chan error: The error channel (currently nil).
func (a *Action) GetErrorChannel() chan error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return nil
}

// GetDoneChannel retrieves the channel used to signal action completion.
// Returns:
//   - chan struct{}: The done channel (currently nil).
func (a *Action) GetDoneChannel() chan struct{} {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return nil
}
