package manager

import (
	"fmt"
	"github.com/faelmori/gastype/internal/actions"
	"github.com/faelmori/gastype/utils"
	"sync"

	l "github.com/faelmori/gastype/log"
	t "github.com/faelmori/gastype/types"
)

// TypeManager manages type-related actions and notifications
type TypeManager struct {
	notifierChan chan string
	email        string
	emailToken   string
	notify       bool
	cfg          t.IConfig
	actions      []t.IAction
	isRunning    bool
	mu           sync.Mutex
}

// NewTypeManager creates a new instance of TypeManager
func NewTypeManager(cfg t.IConfig) t.ITypeManager {
	return &TypeManager{
		cfg:          cfg,
		notifierChan: make(chan string),
		actions:      []t.IAction{},
		isRunning:    false,
	}
}

// Getters
func (tm *TypeManager) GetNotifierChan() chan string { return tm.notifierChan }
func (tm *TypeManager) GetEmail() string             { return tm.email }
func (tm *TypeManager) GetEmailToken() string        { return tm.emailToken }
func (tm *TypeManager) GetNotify() bool              { return tm.notify }
func (tm *TypeManager) GetConfig() t.IConfig         { return tm.cfg }
func (tm *TypeManager) GetActions() []t.IAction      { return tm.actions }
func (tm *TypeManager) IsRunning() bool              { return tm.isRunning }

// Setters
func (tm *TypeManager) SetNotifierChan(notifierChan chan string) { tm.notifierChan = notifierChan }
func (tm *TypeManager) SetEmail(email string)                    { tm.email = email }
func (tm *TypeManager) SetEmailToken(emailToken string)          { tm.emailToken = emailToken }
func (tm *TypeManager) SetNotify(notify bool)                    { tm.notify = notify }
func (tm *TypeManager) SetConfig(cfg t.IConfig)                  { tm.cfg = cfg }
func (tm *TypeManager) AddAction(action t.IAction)               { tm.actions = append(tm.actions, action) }

// Action Management
func (tm *TypeManager) StartChecking(workerCount int) error {
	if len(tm.actions) == 0 {
		return fmt.Errorf("no actions available to execute")
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.isRunning {
		return fmt.Errorf("manager is already running")
	}

	workerManager := NewWorkerManager(workerCount)
	for _, action := range tm.actions {
		if action.CanExecute() {
			workerManager.JobQueue <- action
		} else {
			l.Warn(fmt.Sprintf("Action %s cannot execute", action.GetType()), nil)
		}
	}

	go workerManager.StartWorkers()
	tm.isRunning = true
	return nil
}

func (tm *TypeManager) StopChecking() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !tm.isRunning {
		l.Warn("manager is not running", nil)
		return
	}

	close(tm.notifierChan)
	tm.isRunning = false
	l.Info("TypeManager stopped successfully", nil)
}

// Load and Save Configurations
func (tm *TypeManager) LoadConfig() error {
	if tm.cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}
	return tm.cfg.Load()
}

func (tm *TypeManager) SaveConfig() error {
	if tm.cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}
	return nil // Implement saving logic here
}

// Notification Management
func (tm *TypeManager) CanNotify() bool {
	return tm.notify && tm.notifierChan != nil
}

func (tm *TypeManager) PrepareActions() error {
	parsedFiles, err := utils.ParseFiles(tm.cfg.GetDir())
	if err != nil {
		return fmt.Errorf("error parsing files: %v", err)
	}
	// Criar ações baseadas nos arquivos analisados.
	for pkgName, files := range parsedFiles {
		action := actions.NewTypeCheckAction(pkgName, files, tm.cfg)
		tm.AddAction(action)
	}
	return nil
}
