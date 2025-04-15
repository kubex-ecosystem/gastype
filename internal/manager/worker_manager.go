package manager

import (
	"sync"

	g "github.com/faelmori/gastype/internal/globals"
	t "github.com/faelmori/gastype/types"
	l "github.com/faelmori/logz"
)

// WorkerManager with corrections
type WorkerManager struct {
	// IWorkerManager interface for worker management
	t.IWorkerManager
	// Threading interface for threading
	// Mutex for thread safety
	// SyncGroup for synchronization
	*g.Threading

	Workers        map[int]t.IWorker
	WorkerPool     t.IWorkerPool
	WorkerCount    int
	JobQueue       chan t.IAction
	JobChannel     chan t.IJob
	Results        chan t.IResult
	StopChannel    chan struct{}
	logger         l.Logger
	MonitorChannel chan t.MonitorMessage
}

// NewWorkerManager creates a new instance
func NewWorkerManager(workerCount int, workerPool t.IWorkerPool, logger l.Logger) t.IWorkerManager {
	if workerCount <= 0 {
		workerCount = 1
	}
	return &WorkerManager{
		Threading:   g.NewThreading(),
		WorkerCount: workerCount,
		WorkerPool:  workerPool,
		Workers:     make(map[int]t.IWorker),

		// Initialize channels with a buffer size of 50
		Results:        make(chan t.IResult, 50),
		StopChannel:    make(chan struct{}, 2),
		JobQueue:       make(chan t.IAction, 50),
		JobChannel:     make(chan t.IJob, 50),
		MonitorChannel: make(chan t.MonitorMessage, 50),

		// Initialize the logger
		logger: logger,
	}
}

func (wm *WorkerManager) StartWorkers() {
	var wg sync.WaitGroup
	wm.logger.DebugCtx("Starting workers", map[string]interface{}{"worker_count": wm.WorkerCount})

	for i := 0; i < wm.WorkerCount; i++ {
		wg.Add(1)
		wkr := wm.WorkerPool.GetWorker(i)
		go func(wkr t.IWorker, wg *sync.WaitGroup) {
			defer wg.Done()
			wkr.GetJobQueue()
		}(wkr, &wg)
	}

	wm.logger.DebugCtx("Worker waiting for all workers to start", nil)

	wg.Wait()

	wm.logger.DebugCtx("All workers started", nil)
}
func (wm *WorkerManager) StopWorkers() {
	wm.logger.InfoCtx("Stopping all workers", nil)
	close(wm.StopChannel)
	close(wm.JobQueue)
}

func (wm *WorkerManager) GetWorkerPool() t.IWorkerPool {
	if wm.WorkerPool == nil {
		wm.logger.ErrorCtx("Worker pool not initialized", nil)
		return nil
	}
	return wm.WorkerPool
}
func (wm *WorkerManager) SetWorkerPool(workerPool t.IWorkerPool) {
	if workerPool == nil {
		wm.logger.ErrorCtx("Worker pool is not initialized", nil)
		return
	}
	wm.WorkerPool = workerPool
}

func (wm *WorkerManager) GetLogger() l.Logger {
	if wm.logger == nil {
		wm.logger = l.GetLogger("GasType")
	}
	return wm.logger
}
func (wm *WorkerManager) SetLogger(logger l.Logger) {
	if logger == nil {
		wm.logger = l.GetLogger("GasType")
	} else {
		wm.logger = logger
	}
}

func (wm *WorkerManager) GetStopChannel() chan struct{} {
	if wm.StopChannel == nil {
		wm.logger.ErrorCtx("StopChannel is not initialized", nil)
		return nil
	}
	return wm.StopChannel
}
func (wm *WorkerManager) SetStopChannel(stopChannel chan struct{}) {
	wm.Lock()
	defer wm.Unlock()
	wm.StopChannel = stopChannel
}

func (wm *WorkerManager) GetMonitorChannel() chan t.MonitorMessage {
	if wm.JobQueue == nil {
		wm.logger.ErrorCtx("JobQueue is not initialized", nil)
		return nil
	}
	return wm.MonitorChannel
}
func (wm *WorkerManager) SetMonitorChannel(monitorChannel chan t.MonitorMessage) {
	wm.Lock()
	defer wm.Unlock()
	wm.MonitorChannel = monitorChannel
}

func (wm *WorkerManager) GetWorkerCount() int {
	if wm.WorkerCount <= 0 {
		wm.logger.ErrorCtx("Worker count is not set", nil)
		return 1
	}
	return wm.WorkerCount
}
func (wm *WorkerManager) GetWorkerLimit() int {
	if wm.WorkerCount <= 0 {
		wm.logger.ErrorCtx("Worker limit is not set", nil)
		return 1
	}
	return wm.WorkerCount
}
func (wm *WorkerManager) GetBuffersSize() int {
	if wm.JobQueue == nil {
		wm.logger.ErrorCtx("JobQueue is not initialized", nil)
		return 0
	}
	return cap(wm.JobQueue)
}
func (wm *WorkerManager) AdjustBufferSize(autoSize bool, size int) {
	if autoSize {
		wm.Lock()
		defer wm.Unlock()
		wm.JobQueue = make(chan t.IAction, size)
	} else {
		if size <= 0 {
			size = 2
		}
		wm.Lock()
		defer wm.Unlock()
		wm.JobQueue = make(chan t.IAction, size)
	}
}
func (wm *WorkerManager) SetWorkerLimit(workerLimit int) error {
	if workerLimit <= 0 {
		wm.logger.ErrorCtx("Worker limit is not set", nil)
		return nil
	}
	wm.WorkerCount = workerLimit
	return nil
}
func (wm *WorkerManager) SetWorkerCount(workerCount int) error {
	if workerCount <= 0 {
		wm.logger.ErrorCtx("Worker count is not set", nil)
		return nil
	}
	wm.WorkerCount = workerCount
	return nil
}
