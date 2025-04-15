package manager

import (
	"fmt"
	a "github.com/faelmori/gastype/internal/actions"
	j "github.com/faelmori/gastype/internal/jobs"
	"github.com/faelmori/gastype/internal/manager/workers"
	t "github.com/faelmori/gastype/types"
	"github.com/faelmori/gastype/utils"
	l "github.com/faelmori/logz"
	"go/ast"
	"net/mail"
	"sync"
	"time"
)

type TypeManager struct {
	mu            sync.RWMutex
	notifierChan  chan interface{}
	email         string
	emailToken    string
	notify        bool
	logger        l.Logger
	cfg           t.IConfig
	actions       []t.IAction
	files         []string
	astFiles      []*ast.File
	workerPool    t.IWorkerPool
	workerManager t.IWorkerManager
	isRunning     bool
}

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

func (tm *TypeManager) GetWorkerManager() t.IWorkerManager {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.workerManager == nil {
		tm.logger.ErrorCtx("Worker manager not initialized", nil)
		return nil
	}
	return tm.workerManager
}
func (tm *TypeManager) SetWorkerManager(wm t.IWorkerManager) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if wm == nil {
		tm.logger.ErrorCtx("Worker manager not initialized", nil)
		return
	}
	tm.workerManager = wm
}
func (tm *TypeManager) SetWorkerPool(wp t.IWorkerPool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if wp == nil {
		tm.logger.ErrorCtx("Worker pool not initialized", nil)
		return
	}
	tm.workerPool = wp
}
func (tm *TypeManager) GetWorkerPool() t.IWorkerPool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.workerPool == nil {
		tm.logger.ErrorCtx("Worker pool not initialized", nil)
		return nil
	}
	return tm.workerPool
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
		tm.logger.DebugCtx(fmt.Sprintf("Email parsed successfully: %s", emailParsed), nil)
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
func (tm *TypeManager) SetFiles(files []string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.files = files
}
func (tm *TypeManager) SetLogger(logger l.Logger) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if logger == nil {
		logger = l.GetLogger("GasType")
	}
	tm.logger = logger
}
func (tm *TypeManager) StartChecking(workerCount int, background bool) error {
	lgr := l.GetLogger("GasType")
	if tm.GetLogger() == nil {
		tm.logger = lgr
	}
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

	if background {
		go tm.listenResults()
	} else {
		tm.listenResults()
	}

	return nil
}
func (tm *TypeManager) StopChecking() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !tm.isRunning {
		tm.logger.WarnCtx("TypeManager is not running", nil)
		return
	}

	if tm.workerManager != nil {
		tm.workerManager.GetStopChannel() <- struct{}{}
	}

	if tm.workerPool != nil && tm.workerPool.IsRunning() {
		tm.workerPool.StopWorkers()
	}

	close(tm.notifierChan)

	tm.isRunning = false

	tm.logger.InfoCtx("TypeManager stopped successfully", nil)

	return
}
func (tm *TypeManager) PrepareActions() error {
	if files, err := tm.GetFilesList(true); err != nil {
		tm.logger.ErrorCtx(fmt.Sprintf("Error getting files list: %s", err.Error()), nil)
		return err
	} else {
		tm.files = files
	}

	tm.logger.DebugCtx(fmt.Sprintf("Preparing actions for %d files", len(tm.files)), nil)

	for _, file := range tm.files {
		if file == "" {
			tm.logger.WarnCtx("Empty file name", nil)
			continue
		}
		tm.logger.DebugCtx(fmt.Sprintf("Preparing action for file: %s", file), nil)

		action := a.NewTypeCheckAction(file, tm.cfg, tm.logger)

		tm.AddAction(action)
	}

	tm.logger.DebugCtx(fmt.Sprintf("Actions prepared: %d", len(tm.actions)), nil)

	return nil
}

func (tm *TypeManager) runWorkerManager(workerCount int) {
	tm.logger.DebugCtx("Starting worker manager", nil)
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
	if !tm.workerPool.IsRunning() {
		tm.logger.InfoCtx(fmt.Sprintf("Worker manager starting with %d workers", workerCount), nil)
		if setWorkerCountErr := tm.workerManager.SetWorkerCount(workerCount); setWorkerCountErr != nil {
			tm.logger.ErrorCtx(fmt.Sprintf("Error setting worker count: %s", setWorkerCountErr.Error()), nil)
			return
		}
		go tm.workerPool.StartWorkers()
		time.Sleep(50 * time.Millisecond)
	}
	if !tm.workerPool.IsRunning() {
		tm.logger.ErrorCtx("Worker pool not running", nil)
		return
	}

	for _, action := range tm.actions {
		if action.CanExecute() {
			jb := j.NewJob(action, tm.GetWorkerPool().GetCancelChannel(), nil, tm.logger)
			tm.workerPool.SubmitJob(jb)
			tm.logger.DebugCtx(fmt.Sprintf("Submitting job: %s", action.GetType()), nil)
			if act, ok := action.(t.IAction); ok {
				go tm.watchAction(act)
			}
		}
	}

	// Listen for notifier messages
	go func(tm t.ITypeManager) {
		var wp t.IWorkerPool
		if tm.GetWorkerPool() == nil {
			tm.GetLogger().ErrorCtx("Worker pool not initialized", nil)
			return
		} else {
			wp = tm.GetWorkerPool()
		}
		for {
			select {
			case msg := <-tm.GetNotifierChan():
				tm.GetLogger().DebugCtx(fmt.Sprintf("Received notifier message: %v", msg), nil)
				if tm.CanNotify() {
					tm.GetLogger().DebugCtx(fmt.Sprintf("Sending notifier message: %v", msg), nil)
					tm.GetNotifierChan() <- msg
				}
			case <-wp.GetStopChannel():
				tm.GetLogger().InfoCtx("Stopping worker manager", nil)
				wp.StopWorkers()
				wp.GetCancelChannel() <- struct{}{}
				return
			}
		}
	}(tm)

	tm.isRunning = true
	tm.logger.InfoCtx("Worker manager started successfully", nil)
}
func (tm *TypeManager) listenResults() {
	var wp t.IWorkerPool
	if tm.GetWorkerPool() == nil {
		tm.GetLogger().ErrorCtx("Worker pool not initialized", nil)
		return
	} else {
		wp = tm.GetWorkerPool()
	}
	for {
		select {
		case msg := <-tm.notifierChan:
			tm.logger.DebugCtx(fmt.Sprintf("Received notifier message: %v", msg), nil)
		case <-wp.GetStopChannel():
			tm.logger.InfoCtx("Stopping notification goroutine", nil)
			tm.StopChecking()
			return
		case <-wp.GetJobChannel():
			tm.logger.DebugCtx("Received job channel message", nil)
			// Process job channel messages here
		case <-wp.GetResultChannel():
			tm.logger.DebugCtx("Received result channel message", nil)
			// Process result channel messages here
		case <-wp.GetMonitorChannel():
			tm.logger.DebugCtx("Received monitor channel message", nil)
			// Process monitor channel messages here
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}
func (tm *TypeManager) watchAction(action t.IAction) {
	time.Sleep(50 * time.Millisecond)

	if tm.CanNotify() {
		tm.GetNotifierChan() <- fmt.Sprintf("Action %s started", action.GetType())
	}

	for {
		select {
		case result := <-action.GetResultsChannel():
			tm.logger.DebugCtx(fmt.Sprintf("Received result: %v", result), nil)
			if tm.CanNotify() {
				tm.GetNotifierChan() <- fmt.Sprintf("Action %s completed", action.GetType())
			}
		case err := <-action.GetErrorChannel():
			tm.logger.ErrorCtx(fmt.Sprintf("Error: %s", err.Error()), nil)
			if tm.CanNotify() {
				tm.GetNotifierChan() <- fmt.Sprintf("Action %s failed", action.GetType())
			}
		case <-action.GetCancelChannel():
			tm.logger.InfoCtx("Action cancelled", nil)
			return
		default:
			time.Sleep(50 * time.Millisecond)
		}
		if !action.IsRunning() || action.GetStatus() == fmt.Sprintf("Completed") {
			tm.logger.InfoCtx(fmt.Sprintf("Action %s completed", action.GetType()), nil)
			if tm.CanNotify() {
				tm.GetNotifierChan() <- fmt.Sprintf("Action %s completed", action.GetType())
			}
			return
		}
	}
}
