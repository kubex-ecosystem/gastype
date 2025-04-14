package workers

import (
	"fmt"
	"sync"

	g "github.com/faelmori/gastype/internal/globals"
	t "github.com/faelmori/gastype/types"
	l "github.com/faelmori/logz"
)

// WorkerPool manages workers and their assigned jobs
type WorkerPool struct {
	workerCount    int
	workerLimit    int
	workerPool     map[int]t.IWorker
	jobChannel     chan t.IJob
	resultChannel  chan t.IResult
	monitorChannel chan t.MonitorMessage
	stopChannel    chan struct{}
	logger         l.Logger
	mu             sync.RWMutex
	isRunning      bool
}

// NewWorkerPool creates a new WorkerPool instance
func NewWorkerPool(workerLimit int, logger l.Logger) t.IWorkerPool {
	return &WorkerPool{
		workerLimit:    workerLimit,
		jobChannel:     make(chan t.IJob, 50),
		resultChannel:  make(chan t.IResult, 50),
		monitorChannel: make(chan t.MonitorMessage, 50),
		stopChannel:    make(chan struct{}, 5),
		workerPool:     make(map[int]t.IWorker, workerLimit),
		isRunning:      false,
		logger:         logger,
	}
}

// SubmitJob adds a job to the jobChannel for processing
func (wp *WorkerPool) SubmitJob(job t.IJob) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.jobChannel == nil {
		wp.logger.ErrorCtx("job channel is not initialized", nil)
		return
	}

	select {
	case wp.jobChannel <- job:
		wp.logger.DebugCtx(fmt.Sprintf("Job %s submitted successfully", job.GetID()), nil)
	default:
		wp.logger.WarnCtx(fmt.Sprintf("Job %s could not be submitted, worker limit reached", job.GetID()), nil)
	}
}

// StartWorkers initializes workers and begins processing jobs
func (wp *WorkerPool) StartWorkers() t.IWorkerPool {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.isRunning {
		wp.logger.WarnCtx("WorkerPool is already running", nil)
		return wp
	}

	wp.isRunning = true
	var wg sync.WaitGroup

	for i := 0; i < wp.workerLimit; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			wp.logger.DebugCtx(fmt.Sprintf("Worker %d started", workerID), nil)

			for {
				select {
				case <-wp.stopChannel:
					wp.logger.DebugCtx(fmt.Sprintf("Worker %d stopped", workerID), nil)
					return
				case job, ok := <-wp.jobChannel:
					if !ok {
						wp.logger.WarnCtx(fmt.Sprintf("Worker %d received nil job", workerID), nil)
						return
					}

					if err := job.Execute(); err != nil {
						wp.logger.ErrorCtx(fmt.Sprintf("Error executing job %s: %v", job.GetID(), err), nil)
					} else {
						wp.logger.DebugCtx(fmt.Sprintf("Job %s completed successfully", job.GetID()), nil)

						result := g.NewResult(job.GetID(), "Success ✅", nil)

						//result := t.IResult{
						//	Package: job.GetID(),
						//	Status:  "Success ✅",
						//}
						wp.resultChannel <- result
					}
				}
			}
		}(i)
	}

	wp.logger.DebugCtx("All workers started successfully", nil)
	wp.logger.DebugCtx(fmt.Sprintf("WorkerPool started with %d workers", wp.workerLimit), nil)
	wp.logger.DebugCtx(fmt.Sprintf("WorkerPool is running: %t", wp.isRunning), nil)

	// Wait for all workers to finish
	wg.Wait()

	wp.logger.DebugCtx("All workers finished processing", nil)
	wp.isRunning = false

	return wp
}

// StopWorkers stops all workers and closes necessary channels
func (wp *WorkerPool) StopWorkers() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if !wp.isRunning {
		wp.logger.WarnCtx("WorkerPool is not running", nil)
		return
	}

	close(wp.stopChannel)
	close(wp.jobChannel)
	wp.isRunning = false
	wp.logger.DebugCtx("WorkerPool stopped successfully", nil)
}

// Getters

func (wp *WorkerPool) GetBuffersSize() int {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	if wp.jobChannel == nil || wp.resultChannel == nil || wp.monitorChannel == nil {
		wp.logger.ErrorCtx("One of the channels is not initialized", nil)
		return 0
	} else {
		wp.logger.DebugCtx("Getting buffer size", nil)
		wp.logger.DebugCtx(fmt.Sprintf("Job channel size: %d", len(wp.jobChannel)), nil)
		wp.logger.DebugCtx(fmt.Sprintf("Result channel size: %d", len(wp.resultChannel)), nil)
		wp.logger.DebugCtx(fmt.Sprintf("Monitor channel size: %d", len(wp.monitorChannel)), nil)
	}
	return len(wp.jobChannel) + len(wp.resultChannel) + len(wp.monitorChannel)
}
func (wp *WorkerPool) AjustBufferSize(autoSize bool, size int) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if autoSize {
		wp.logger.DebugCtx("Auto-sizing buffer size", nil)
		cpuCount := g.NewEnvironment().CpuCount()
		memTotal := g.NewEnvironment().MemTotal()
		// Calculate buffer size based on CPU count and memory
		// For example, we can set the buffer size to 2 * CPU count
		// and limit it to a maximum of 100
		size = 2 * cpuCount
		memTotal = memTotal / 1024 / 1024 // Convert to MB
		expectSize := size * 10
		lowerBound := expectSize - cpuCount
		if lowerBound > memTotal {
			lowerBound = memTotal
		} else if lowerBound < 2 {
			lowerBound = 2
		}
		if size > 100 {
			wp.logger.ErrorCtx("Buffer size exceeds maximum limit of 100, limiting to the allowed size", nil)
			size = 100
		}
		if expectSize > memTotal {
			wp.logger.ErrorCtx("Buffer size exceeds available memory, setting to memory limit", nil)
			size = memTotal
		}
		if size < lowerBound {
			wp.logger.ErrorCtx("Buffer size is too small, setting to lower bound", nil)
			size = lowerBound
		}

		wp.jobChannel = make(chan t.IJob, size)
		wp.resultChannel = make(chan t.IResult, size)
		wp.monitorChannel = make(chan t.MonitorMessage, size)
	} else {
		wp.logger.DebugCtx("Setting buffer size manually", nil)
		wp.jobChannel = make(chan t.IJob, size)
		wp.resultChannel = make(chan t.IResult, size)
		wp.monitorChannel = make(chan t.MonitorMessage, size)
	}
}
func (wp *WorkerPool) GetStopChannel() chan struct{} {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return wp.stopChannel
}
func (wp *WorkerPool) GetLogger() l.Logger {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return wp.logger
}
func (wp *WorkerPool) GetMonitorChannel() chan t.MonitorMessage {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return wp.monitorChannel
}
func (wp *WorkerPool) GetJobChannel() chan t.IJob {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return wp.jobChannel
}
func (wp *WorkerPool) GetResultChannel() chan t.IResult {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return wp.resultChannel
}
func (wp *WorkerPool) GetWorkerCount() int {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return len(wp.workerPool)
}
func (wp *WorkerPool) GetWorkerLimit() int {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return wp.workerLimit
}
func (wp *WorkerPool) GetWorkerPool() map[int]t.IWorker {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return wp.workerPool
}
func (wp *WorkerPool) IsRunning() bool {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return wp.isRunning
}

// Setters

func (wp *WorkerPool) SetWorkerLimit(workerLimit int) error {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	if workerLimit <= 0 {
		return fmt.Errorf("worker limit must be greater than 0")
	} else if workerLimit > 100 {
		return fmt.Errorf("worker limit must be less than or equal to 100")
	} else if workerLimit < wp.workerCount {
		return fmt.Errorf("worker limit must be greater than current worker count")
	} else {
		wp.workerLimit = workerLimit
		return nil
	}
}
