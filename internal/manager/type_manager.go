package manager

import (
	"encoding/json"
	"fmt"
	"github.com/faelmori/gastype/internal/actions"
	"github.com/faelmori/gastype/utils"
	"go/ast"
	"os"
	"sync"

	t "github.com/faelmori/gastype/types"
	l "github.com/faelmori/logz"
)

// TypeManager manages type-related actions and notifications
type TypeManager struct {
	notifierChan chan interface{}
	email        string
	emailToken   string
	notify       bool

	logger  l.Logger
	cfg     t.IConfig
	actions []t.IAction

	files    []string
	astFiles []*ast.File

	workerManager t.IWorker
	isRunning     bool
	mu            sync.Mutex
}

// NewTypeManager creates a new instance of TypeManager
func NewTypeManager(cfg t.IConfig, logger l.Logger) t.ITypeManager {
	if logger == nil {
		logger = l.GetLogger("GasType")
	}
	return &TypeManager{
		cfg:           cfg,
		logger:        logger,
		files:         make([]string, 0),
		astFiles:      make([]*ast.File, 0),
		notifierChan:  make(chan interface{}, 20),
		email:         "",
		emailToken:    "",
		notify:        false,
		workerManager: nil,
		actions:       []t.IAction{},
		isRunning:     false,
	}
}

// Getters
func (tm *TypeManager) GetNotifierChan() chan interface{} { return tm.notifierChan }
func (tm *TypeManager) GetEmail() string                  { return tm.email }
func (tm *TypeManager) GetEmailToken() string             { return tm.emailToken }
func (tm *TypeManager) GetNotify() bool                   { return tm.notify }
func (tm *TypeManager) GetConfig() t.IConfig              { return tm.cfg }
func (tm *TypeManager) GetActions() []t.IAction           { return tm.actions }
func (tm *TypeManager) GetLogger() l.Logger               { return tm.logger }
func (tm *TypeManager) GetFilesList(force bool) ([]string, []*ast.File, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if !force {
		if tm.files != nil && len(tm.files) > 0 {
			return tm.files, tm.astFiles, nil
		}
	}
	files := make([]string, 0)
	astFiles := make([]*ast.File, 0)

	if collectErr := utils.CollectGoFiles(tm.cfg.GetDir(), &files, &astFiles, tm.logger); collectErr != nil {
		tm.logger.ErrorCtx(fmt.Sprintf("Error collecting Go files: %s", collectErr.Error()), nil)
		return nil, nil, collectErr
	}
	if len(files) == 0 {
		tm.logger.WarnCtx("No Go files found", nil)
		return nil, nil, fmt.Errorf("no Go files found")
	}
	tm.logger.DebugCtx(fmt.Sprintf("Collected %d Go files", len(files)), nil)
	return files, astFiles, nil
}
func (tm *TypeManager) IsRunning() bool { return tm.isRunning }

// Setters
func (tm *TypeManager) SetNotifierChan(notifierChan chan interface{}) { tm.notifierChan = notifierChan }
func (tm *TypeManager) SetEmail(email string)                         { tm.email = email }
func (tm *TypeManager) SetEmailToken(emailToken string)               { tm.emailToken = emailToken }
func (tm *TypeManager) SetNotify(notify bool)                         { tm.notify = notify }
func (tm *TypeManager) SetConfig(cfg t.IConfig)                       { tm.cfg = cfg }
func (tm *TypeManager) AddAction(action t.IAction)                    { tm.actions = append(tm.actions, action) }

// StartChecking begins the process of checking Go files
func (tm *TypeManager) StartChecking(workerCount int) error {
	lgr := l.GetLogger("GasType")
	if tm.logger == nil {
		tm.logger = lgr
	}

	if tm.workerManager != nil {
		if jq := tm.workerManager.GetJobQueue(); len(jq) > 0 {
			tm.logger.WarnCtx(fmt.Sprintf("Job queue is not empty, %d jobs pending", len(jq)), nil)
			return fmt.Errorf("job queue is not empty, %d jobs pending", len(jq))
		}
		tm.workerManager.StopWorkers()
	}

	tm.logger.InfoCtx("Starting worker manager", nil)
	tm.workerManager = NewWorkerManager(workerCount, tm.logger)
	go tm.workerManager.StartWorkers()

	tm.logger.InfoCtx("Worker manager started", nil)
	if prepareErr := tm.PrepareActions(); prepareErr != nil {
		tm.logger.ErrorCtx(fmt.Sprintf("Error preparing actions: %s", prepareErr.Error()), nil)
		return prepareErr
	}

	tm.logger.InfoCtx(fmt.Sprintf("Prepared %d actions", len(tm.actions)), nil)
	if len(tm.actions) == 0 {
		tm.logger.WarnCtx("no actions available to execute", nil)
		return fmt.Errorf("no actions available to execute")
	}

	tm.logger.InfoCtx("Preparing actions", nil)
	if tm.isRunning {
		tm.logger.WarnCtx("manager is already running", nil)
		return fmt.Errorf("manager is already running")
	}

	tm.logger.InfoCtx(fmt.Sprintf("Loading files from %s", tm.cfg.GetDir()), nil)
	//if err := tm.LoadConfig(); err != nil {
	//	tm.logger.ErrorCtx(fmt.Sprintf("Error loading configuration: %s", err.Error()), nil)
	//	return err
	//}

	// Save configuration
	//if err := tm.SaveConfig(); err != nil {
	//	tm.logger.ErrorCtx(fmt.Sprintf("Error saving configuration: %s", err.Error()), nil)
	//	return err
	//}

	//tm.mu.Lock()
	//defer tm.mu.Unlock()

	tm.logger.InfoCtx("Starting ReportNotifier", nil)
	reportManager := NewReportManager[interface{}](tm.logger)
	tm.SetNotifierChan(reportManager.notifyChan)

	// Start the report manager
	go func(ch chan interface{}, lgr l.Logger) {
		l.GetLogger("GasType")
		lgr.DebugCtx("Starting report manager", nil)
		for {
			select {
			case report := <-ch:
				if report != nil {
					lgr.DebugCtx(fmt.Sprintf("Report: %v", report), nil)
					// Handle the report here
				} else {
					lgr.WarnCtx("Received nil report", nil)
				}
			}
		}
	}(reportManager.notifyChan, tm.logger)

	tm.logger.InfoCtx(fmt.Sprintf("Worker manager starting with %d workers", workerCount), nil)
	wm := tm.workerManager

	// Start the worker manager
	go func(ch chan t.IAction, lgr l.Logger, tm *TypeManager) {
		lgr.InfoCtx("Starting action job queue", nil)
		for {
			select {
			case job := <-ch:
				if job != nil {
					lgr.DebugCtx(fmt.Sprintf("Executing job: %s", job.GetType()), nil)
					if err := job.Execute(); err != nil {
						lgr.ErrorCtx(fmt.Sprintf("Error executing job: %s", err.Error()), nil)
					} else {
						lgr.DebugCtx(fmt.Sprintf("Job %s executed successfully", job.GetType()), nil)
						if tm.notify {
							lgr.DebugCtx("Sending notification", nil)
							// Send notification logic here
						} else {
							lgr.WarnCtx("Notification is disabled", nil)
						}
						// Optionally log the job results:
						res := job.GetResults()
						if len(res) > 0 {
							lgr.DebugCtx(fmt.Sprintf("Job %s returned results", job.GetType()), nil)
						} else {
							lgr.WarnCtx(fmt.Sprintf("Job %s has no results", job.GetType()), nil)
						}
					}
				} else {
					lgr.WarnCtx("Received nil job", nil)
				}
			case <-tm.notifierChan:
				lgr.DebugCtx("Notification received", nil)
			}
		}
	}(wm.GetJobQueue(), tm.logger, tm)

	// Execute actions
	for _, action := range tm.actions {
		if action.CanExecute() {
			tm.logger.DebugCtx(fmt.Sprintf("Action %s is executing", action.GetType()), nil)
			wm.GetJobQueue() <- action
		} else {
			tm.logger.WarnCtx(fmt.Sprintf("Action %s cannot execute", action.GetType()), nil)
		}
	}

	tm.isRunning = true

	return nil
}
func (tm *TypeManager) StopChecking() {
	lgr := l.GetLogger("GasType")
	if tm.logger == nil {
		tm.logger = lgr
	}

	if !tm.isRunning {
		tm.logger.WarnCtx("manager is not running", nil)
		return
	}

	close(tm.notifierChan)
	// Save configuration/results before stopping
	if err := tm.SaveConfig(); err != nil {
		tm.logger.ErrorCtx(fmt.Sprintf("Error saving configuration: %v", err), nil)
	}
	tm.isRunning = false
	tm.logger.DebugCtx("TypeManager stopped successfully", nil)
}
func (tm *TypeManager) LoadConfig() error {
	if tm.cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}
	return tm.cfg.Load()
}
func (tm *TypeManager) SaveConfig() error {
	// ...existing code before saving...
	if tm.cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}
	// Collect results from all actions
	results := make([]map[string]interface{}, 0)
	for _, act := range tm.actions {
		if res := act.GetResults(); res != nil && len(res) > 0 {
			results = append(results, res)
		}
	}
	dataBytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		tm.logger.ErrorCtx(fmt.Sprintf("Error marshalling results: %v", err), nil)
		return err
	}
	// Get the output file from config
	outFile := tm.cfg.GetOutputFile()
	tm.logger.InfoCtx(fmt.Sprintf("Saving results to file: %s", outFile), nil)
	if writeErr := os.WriteFile(outFile, dataBytes, 0644); writeErr != nil {
		tm.logger.ErrorCtx(fmt.Sprintf("Error writing results to file: %v", writeErr), nil)
		return writeErr
	}
	tm.logger.InfoCtx(fmt.Sprintf("Results successfully saved to %s", outFile), nil)
	return nil
}
func (tm *TypeManager) CanNotify() bool {
	return tm.notify && tm.notifierChan != nil
}
func (tm *TypeManager) PrepareActions() error {
	//tm.mu.Lock()
	//defer tm.mu.Unlock()
	tm.logger.InfoCtx("Preparing actions", nil)
	if files, astFiles, filesErr := tm.GetFilesList(true); filesErr != nil {
		tm.logger.ErrorCtx(fmt.Sprintf("Error getting files list: %s", filesErr.Error()), nil)
		return filesErr
	} else {
		tm.files = files
		tm.astFiles = astFiles
		if len(tm.files) == 0 {
			tm.logger.WarnCtx("No parsed files found", nil)
			return fmt.Errorf("no parsed files found")
		}
		tm.logger.InfoCtx(fmt.Sprintf("Found %d parsed files", len(tm.files)), map[string]interface{}{})
		for _, file := range tm.astFiles {
			tm.logger.InfoCtx(fmt.Sprintf("Preparing action for file: %s", file.Name.Name), nil)
			action := actions.NewTypeCheckAction(file.Name.Name, []*ast.File{file}, tm.cfg, tm.logger)
			tm.AddAction(action)
		}
		if len(tm.actions) == 0 {
			tm.logger.WarnCtx("No actions prepared", nil)
			return fmt.Errorf("no actions prepared")
		}
		tm.logger.InfoCtx(fmt.Sprintf("Prepared %d actions", len(tm.actions)), nil)
	}
	return nil
}
