package manager

import (
	"fmt"
	a "github.com/faelmori/gastype/internal/actions"
	"github.com/faelmori/gastype/internal/manager/workers"
	"github.com/faelmori/gastype/utils"
	"go/ast"
	"net/mail"
	"sync"

	t "github.com/faelmori/gastype/types"
	l "github.com/faelmori/logz"
)

// TypeManager with corrections
type TypeManager struct {
	notifierChan  chan interface{}
	email         string
	emailToken    string
	notify        bool
	logger        l.Logger
	cfg           t.IConfig
	actions       []t.IAction
	files         []string
	astFiles      []*ast.File
	workerManager t.IWorker
	workerPool    t.IWorkerPool
	isRunning     bool
	mu            sync.RWMutex
}

// NewTypeManager creates a new instance
func NewTypeManager(cfg t.IConfig, logger l.Logger) t.ITypeManager {
	if logger == nil {
		if cfg.GetLogger() != nil {
			logger = l.GetLogger("GasType")
		} else {
			logger = cfg.GetLogger()
		}
	}

	workerManager := NewWorkerManager(cfg.GetWorkerLimit(), workers.NewWorkerPool(cfg.GetWorkerLimit(), logger), logger)

	return &TypeManager{
		mu:            sync.RWMutex{},
		cfg:           cfg,
		logger:        logger,
		workerPool:    workerManager.GetWorkerPool(),
		workerManager: workerManager,
		email:         "",
		emailToken:    "",
		notify:        false,
		actions:       make([]t.IAction, 0),
		files:         []string{},
		astFiles:      []*ast.File{},
		notifierChan:  make(chan interface{}, 20),
		isRunning:     false,
	}
}

// Getters

func (tm *TypeManager) GetStopChannel() chan struct{} {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	wp := tm.workerManager.GetWorkerPool()
	if wp == nil {
		tm.logger.ErrorCtx("Worker pool not initialized", nil)
		return nil
	}
	return wp.GetStopChannel()
}
func (tm *TypeManager) GetNotifierChan() chan interface{} {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.notifierChan
}
func (tm *TypeManager) GetEmail() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.email
}
func (tm *TypeManager) GetEmailToken() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.emailToken
}
func (tm *TypeManager) GetNotify() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.notify
}
func (tm *TypeManager) GetConfig() t.IConfig {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.cfg
}
func (tm *TypeManager) GetActions() []t.IAction {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.actions
}
func (tm *TypeManager) GetLogger() l.Logger {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.logger == nil {
		tm.logger = l.GetLogger("GasType")
	}

	return tm.logger
}
func (tm *TypeManager) GetFilesList(force bool) ([]string, error) {
	//tm.mu.Lock()
	//defer tm.mu.Unlock()

	if !force {
		if tm.files != nil && len(tm.files) > 0 {
			return tm.files, nil
		}
	}
	files := make([]string, 0)

	if collectErr := utils.CollectGoFiles(tm.cfg.GetDir(), &files, tm.logger); collectErr != nil {
		tm.logger.ErrorCtx(fmt.Sprintf("Error collecting Go files: %s", collectErr.Error()), nil)
		return nil, collectErr
	}
	if len(files) == 0 {
		tm.logger.WarnCtx("No Go files found", nil)
		return nil, fmt.Errorf("no Go files found")
	}

	tm.logger.NoticeCtx(fmt.Sprintf("Collected %d Go files", len(files)), nil)
	return files, nil
}
func (tm *TypeManager) IsRunning() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.isRunning
}

// Setters

func (tm *TypeManager) SetNotifierChan(notifierChan chan interface{}) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.notifierChan = notifierChan
}
func (tm *TypeManager) SetEmail(email string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if email == "" {
		tm.logger.ErrorCtx("Email is empty", nil)
		return
	}
	if emailParsed, err := mail.ParseAddress(email); err != nil {
		tm.logger.ErrorCtx(fmt.Sprintf("Invalid email format: %s", err.Error()), nil)
		return
	} else {
		tm.logger.NoticeCtx(fmt.Sprintf("Email parsed successfully: %s", emailParsed), nil)
		tm.email = emailParsed.String()
	}
}
func (tm *TypeManager) SetEmailToken(emailToken string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	// Validate email token
	if emailToken == "" {
		tm.logger.ErrorCtx("Email token is empty", nil)
		return
	}
	if len(emailToken) < 10 {
		tm.logger.ErrorCtx("Email token is too short", nil)
		return
	}
	if len(emailToken) > 100 {
		tm.logger.ErrorCtx("Email token is too long", nil)
		return
	}
	if tm.email == "" {
		tm.logger.ErrorCtx("Email is not set", nil)
		return
	}
	tm.emailToken = emailToken
}
func (tm *TypeManager) SetNotify(notify bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.notify = notify
}
func (tm *TypeManager) SetConfig(cfg t.IConfig) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.cfg = cfg
}
func (tm *TypeManager) AddAction(action t.IAction) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if tm.actions == nil {
		tm.logger.WarnCtx("Actions list is nil, initializing", nil)
		tm.actions = make([]t.IAction, 0)
	}
	if action == nil {
		tm.logger.WarnCtx("Action is nil, cannot add", nil)
		return
	}
	tm.logger.DebugCtx(fmt.Sprintf("Adding action: %s", action.GetType()), nil)
	tm.actions = append(tm.actions, action)
}

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
func (tm *TypeManager) CanNotify() bool {
	return tm.notify && tm.notifierChan != nil
}

func (tm *TypeManager) runWorkerManager(workerCount int) {
	tm.logger.InfoCtx("Starting worker manager", nil)
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.workerPool == nil {
		tm.workerPool = workers.NewWorkerPool(workerCount, tm.logger)
	}
	if tm.workerManager == nil {
		tm.workerManager = NewWorkerManager(tm.cfg.GetWorkerLimit(), tm.workerPool, tm.logger)
	}
	if tm.workerPool.GetJobChannel() == nil {
		tm.workerPool.SetJobChannel(make(chan t.IJob, 50))
	}
	if tm.workerManager.GetJobQueue() == nil {
		tm.workerManager.SetJobQueue(make(chan t.IAction, 50))
	}

	tm.logger.InfoCtx(fmt.Sprintf("Worker manager starting with %d workers", workerCount), nil)
	go tm.workerManager.StartWorkers()

	for _, action := range tm.actions {
		if action.CanExecute() {
			tm.workerManager.GetJobQueue() <- action
		}
	}

	// Listen for notifier messages
	go func(tm *TypeManager) {
		if tm.workerPool == nil {
			tm.logger.ErrorCtx("Worker pool not initialized", nil)
			return
		}
		for {
			select {
			case msg := <-tm.notifierChan:
				tm.logger.NoticeCtx(fmt.Sprintf("Received notifier message: %v", msg), nil)
			case <-tm.GetStopChannel():
				tm.logger.InfoCtx("Stopping worker manager notification goroutine", nil)
				return
			}
		}
	}(tm)
}
func (tm *TypeManager) SetFiles(files []string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.files = files
}
func (tm *TypeManager) SetWorkerManager(wm t.IWorker) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if wm == nil {
		tm.logger.ErrorCtx("Worker manager is nil", nil)
		return
	}
	if wm.GetWorkerPool() == nil {
		tm.logger.ErrorCtx("Worker pool is nil", nil)
		return
	}

	tm.workerManager = wm
}
func (tm *TypeManager) SetWorkerPool(wp t.IWorkerPool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if wp == nil {
		tm.logger.ErrorCtx("Worker pool is nil", nil)
		return
	}

	tm.workerPool = wp
}
func (tm *TypeManager) SetLogger(logger l.Logger) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if logger == nil {
		logger = l.GetLogger("GasType")
	}
	tm.logger = logger
}

func (tm *TypeManager) GetWorkerManager() t.IWorker {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.workerManager == nil {
		tm.logger.ErrorCtx("Getting Worker manager not initialized", nil)
		return nil
	}

	return tm.workerManager
}
func (tm *TypeManager) GetWorkerPool() t.IWorkerPool {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.workerPool == nil {
		tm.logger.ErrorCtx("Worker pool not initialized", nil)
		return nil
	}

	return tm.workerManager.GetWorkerPool()
}

func (tm *TypeManager) StartChecking(workerCount int) error {
	lgr := l.GetLogger("GasType")
	if tm.GetLogger() == nil {
		tm.logger = lgr
	}

	//tm.mu.Lock()
	//defer tm.mu.Unlock()

	if tm.IsRunning() {
		tm.logger.WarnCtx("TypeManager already running", nil)
		return fmt.Errorf("manager already running")
	}

	tm.SetWorkerPool(workers.NewWorkerPool(workerCount, tm.logger))
	tm.SetWorkerManager(NewWorkerManager(workerCount, tm.GetWorkerPool(), tm.logger))

	if prepareErr := tm.PrepareActions(); prepareErr != nil {
		tm.logger.ErrorCtx(fmt.Sprintf("Error preparing actions: %s", prepareErr.Error()), nil)
		return prepareErr
	}

	go tm.runWorkerManager(workerCount)

	if !tm.IsRunning() {
		tm.logger.InfoCtx("Starting worker manager", nil)
		tm.SetWorkerManager(tm.GetWorkerManager().StartWorkers())
	}

	if tm.IsRunning() {
		for _, action := range tm.GetActions() {
			if action.CanExecute() {
				tm.GetWorkerManager().GetJobQueue() <- action
			} else {
				tm.logger.WarnCtx(fmt.Sprintf("Action %s cannot be executed", action.GetErrors()), nil)
			}
		}
	} else {
		tm.logger.ErrorCtx("Starting Worker manager not initialized", nil)
		return fmt.Errorf("worker manager not initialized")
	}

	tm.isRunning = true
	return nil
}
func (tm *TypeManager) StopChecking() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !tm.isRunning {
		tm.logger.WarnCtx("TypeManager is not running", nil)
		return
	}

	tm.workerManager.StopWorkers()
	close(tm.notifierChan)
	tm.isRunning = false
	tm.logger.InfoCtx("TypeManager stopped successfully", nil)
}

func (tm *TypeManager) PrepareActions() error {
	if files, err := tm.GetFilesList(true); err != nil {
		tm.logger.ErrorCtx(fmt.Sprintf("Error getting files list: %s", err.Error()), nil)
		return err
	} else {
		tm.files = files
	}

	tm.logger.NoticeCtx(fmt.Sprintf("Preparing actions for %d files", len(tm.files)), nil)

	for _, file := range tm.files {
		if file == "" {
			tm.logger.WarnCtx("Empty file name", nil)
			continue
		}
		tm.logger.NoticeCtx(fmt.Sprintf("Preparing action for file: %s", file), nil)

		action := a.NewTypeCheckAction(tm.files, tm.cfg, tm.logger)

		tm.AddAction(action)

		//tm.logger.NoticeCtx(fmt.Sprintf("Submitting job for file: %s", file), nil)
		//tm.workerManager.GetJobQueue() <- action

		//tm.logger.NoticeCtx(fmt.Sprintf("Preparing job for file: %s", file), nil)
		//tm.workerPool.SubmitJob(j.NewJob(action))
	}

	tm.logger.NoticeCtx(fmt.Sprintf("Actions prepared: %d", len(tm.actions)), nil)

	return nil
}
