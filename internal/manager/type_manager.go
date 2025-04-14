package manager

import (
	"fmt"
	a "github.com/faelmori/gastype/internal/actions"
	j "github.com/faelmori/gastype/internal/jobs"
	"github.com/faelmori/gastype/internal/manager/workers"
	"github.com/faelmori/gastype/utils"
	"go/ast"
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
	mu            sync.Mutex
}

// NewTypeManager creates a new instance
func NewTypeManager(cfg t.IConfig, logger l.Logger) t.ITypeManager {
	if logger == nil {
		logger = l.GetLogger("GasType")
	}
	return &TypeManager{
		cfg:          cfg,
		logger:       logger,
		files:        []string{},
		astFiles:     []*ast.File{},
		notifierChan: make(chan interface{}, 20),
		isRunning:    false,
	}
}

// Getters

func (tm *TypeManager) GetStopChannel() chan struct{} {
	wp := tm.workerManager.GetWorkerPool()
	if wp == nil {
		tm.logger.ErrorCtx("Worker pool not initialized", nil)
		return nil
	}
	return wp.GetStopChannel()
}
func (tm *TypeManager) GetNotifierChan() chan interface{} { return tm.notifierChan }
func (tm *TypeManager) GetEmail() string                  { return tm.email }
func (tm *TypeManager) GetEmailToken() string             { return tm.emailToken }
func (tm *TypeManager) GetNotify() bool                   { return tm.notify }
func (tm *TypeManager) GetConfig() t.IConfig              { return tm.cfg }
func (tm *TypeManager) GetActions() []t.IAction           { return tm.actions }
func (tm *TypeManager) GetLogger() l.Logger               { return tm.logger }
func (tm *TypeManager) GetFilesList(force bool) ([]string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
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
func (tm *TypeManager) IsRunning() bool { return tm.isRunning }

// Setters

func (tm *TypeManager) SetNotifierChan(notifierChan chan interface{}) { tm.notifierChan = notifierChan }
func (tm *TypeManager) SetEmail(email string)                         { tm.email = email }
func (tm *TypeManager) SetEmailToken(emailToken string)               { tm.emailToken = emailToken }
func (tm *TypeManager) SetNotify(notify bool)                         { tm.notify = notify }
func (tm *TypeManager) SetConfig(cfg t.IConfig)                       { tm.cfg = cfg }
func (tm *TypeManager) AddAction(action t.IAction)                    { tm.actions = append(tm.actions, action) }

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

func (tm *TypeManager) StartChecking(workerCount int) error {
	lgr := l.GetLogger("GasType")
	if tm.logger == nil {
		tm.logger = lgr
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.isRunning {
		tm.logger.WarnCtx("TypeManager already running", nil)
		return fmt.Errorf("manager already running")
	}

	wp := workers.NewWorkerPool(workerCount, tm.logger)

	if tm.workerManager == nil {
		tm.workerManager = NewWorkerManager(workerCount, wp, tm.logger)
	}

	if prepareErr := tm.PrepareActions(); prepareErr != nil {
		tm.logger.ErrorCtx(fmt.Sprintf("Error preparing actions: %s", prepareErr.Error()), nil)
		return prepareErr
	}

	go tm.runWorkerManager(workerCount)

	if tm.IsRunning() {
		for _, action := range tm.actions {
			if action.CanExecute() {
				tm.workerManager.GetJobQueue() <- action
			} else {
				tm.logger.WarnCtx(fmt.Sprintf("Action %s cannot be executed", action.GetErrors()), nil)
			}
		}
	} else {
		tm.logger.ErrorCtx("Worker manager not initialized", nil)
		return fmt.Errorf("worker manager not initialized")
	}

	tm.isRunning = true
	return nil
}
func (tm *TypeManager) PrepareActions() error {
	_, err := tm.GetFilesList(true)
	if err != nil {
		tm.logger.ErrorCtx(fmt.Sprintf("Error getting files list: %s", err.Error()), nil)
		return err
	}

	actions := make([]t.IAction, 0)
	wm := tm.GetWorkerManager()
	if wm == nil {
		wp := workers.NewWorkerPool(tm.cfg.GetWorkerLimit(), tm.logger)
		tm.SetWorkerManager(NewWorkerManager(tm.cfg.GetWorkerLimit(), wp, tm.logger))
	} else {
		tm.workerManager = wm
	}
	if tm.workerManager == nil {
		tm.logger.ErrorCtx("Worker manager not initialized", nil)
		return fmt.Errorf("worker manager not initialized")
	}
	wp := tm.workerManager.GetWorkerPool()
	if wp == nil {
		tm.logger.ErrorCtx("Worker pool not initialized", nil)
		return fmt.Errorf("worker pool not initialized")
	} else {
		tm.workerPool = wp
	}
	if !tm.workerPool.IsRunning() {
		tm.logger.InfoCtx("Worker pool and workers not running, starting now", nil)
		go tm.workerPool.StartWorkers()
	}
	jbQueue := tm.workerManager.GetJobQueue()
	if jbQueue == nil {
		jbQueue = make(chan t.IAction, 50)
		tm.workerManager.SetJobQueue(jbQueue)
	}
	jbCh := tm.workerPool.GetJobChannel()
	if jbCh == nil {
		jbCh = make(chan t.IJob, 50)
		tm.workerPool.SetJobChannel(jbCh)
	}

	for _, file := range tm.files {
		if file == "" {
			tm.logger.WarnCtx("Empty file name", nil)
			continue
		}
		tm.logger.NoticeCtx(fmt.Sprintf("Preparing action for file: %s", file), nil)
		action := a.NewTypeCheckAction(tm.files, tm.cfg, tm.logger)
		if jbQueue := tm.workerManager.GetJobQueue(); jbQueue != nil {
			jbQueue <- action
			tm.workerPool.SubmitJob(j.NewJob(action))
		} else {
			tm.logger.ErrorCtx("Job queue not initialized", nil)
			return fmt.Errorf("job queue not initialized")
		}
		actions = append(actions, action)
	}

	tm.SetActions(actions)

	if len(actions) == 0 {
		if wm == nil && tm.workerManager == nil {
			tm.logger.ErrorCtx("Worker manager not initialized", nil)
			return fmt.Errorf("worker manager not initialized")
		} else {
			tm.logger.ErrorCtx("Stopping workers", nil)
			tm.workerManager.StopWorkers()
		}
		tm.logger.WarnCtx(fmt.Sprintf("No actions prepared: %d", len(tm.GetActions())), nil)
		return fmt.Errorf("no actions prepared")
	}
	return nil
}
func (tm *TypeManager) runWorkerManager(workerCount int) {
	tm.logger.InfoCtx(fmt.Sprintf("Worker manager starting with %d workers", workerCount), nil)
	tm.workerManager.StartWorkers()

	for _, action := range tm.actions {
		if action.CanExecute() {
			tm.workerManager.GetJobQueue() <- action
		}
	}

	// Listen for notifier messages
	go func(ch chan interface{}) {
		wp := tm.workerManager.GetWorkerPool()
		if wp == nil {
			tm.logger.ErrorCtx("Worker pool not initialized", nil)
			return
		}
		stopChan := wp.GetStopChannel()
		for {
			select {
			case msg := <-ch:
				tm.logger.NoticeCtx(fmt.Sprintf("Received notifier message: %v", msg), nil)
			case <-stopChan:
				tm.logger.InfoCtx("Stopping worker manager notification goroutine", nil)
				return
			}
		}
	}(tm.notifierChan)
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
func (tm *TypeManager) SetFiles(files []string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.files = files
}
func (tm *TypeManager) SetWorkerManager(wm t.IWorker) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.workerManager = wm
}
func (tm *TypeManager) SetLogger(logger l.Logger) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if logger == nil {
		logger = l.GetLogger("GasType")
	}
	tm.logger = logger
}
func (tm *TypeManager) SetActions(actions []t.IAction) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.actions = actions
}
func (tm *TypeManager) GetWorkerManager() t.IWorker {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.workerManager
}
func (tm *TypeManager) GetWorkerPool() t.IWorkerPool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if tm.workerManager == nil {
		return nil
	}
	return tm.workerManager.GetWorkerPool()
}
