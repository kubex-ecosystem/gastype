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
	t.IWorkerPool
	*g.Threading

	workerCount    int
	workerLimit    int
	workerPool     map[int]t.IWorker
	jobChannel     chan t.IJob
	cancelChannel  chan struct{}
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
		Threading:      g.NewThreading(),
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

func (wp *WorkerPool) StartWorkers() t.IWorkerPool {
	wp.Lock()
	defer wp.Unlock()

	if wp.isRunning {
		wp.logger.WarnCtx("WorkerPool is already running", nil)
		return wp
	}

	wp.isRunning = true
	var wg sync.WaitGroup

	for i := 0; i < wp.workerLimit; i++ {
		wp.Add(1)

		go func(worker t.IWorker, dfer func()) {
			if dfer != nil {
				defer dfer()
			}
			wp.logger.NoticeCtx(fmt.Sprintf("Worker %d started", worker.GetID()), nil)

			for {
				select {
				case <-worker.GetStopChannel():
					wp.logger.NoticeCtx(fmt.Sprintf("Worker %d stopped", worker.GetID()), nil)
					return
				case job, ok := <-worker.GetJobQueue():
					if !ok {
						wp.logger.WarnCtx(fmt.Sprintf("Worker %d received nil job", worker.GetID()), nil)
						return
					}

					if err := job.Execute(); err != nil {
						wp.logger.ErrorCtx(fmt.Sprintf("Error executing job %s: %v", job.GetID(), err), nil)
					} else {
						wp.logger.NoticeCtx(fmt.Sprintf("Job %s completed successfully", job.GetID()), nil)
					}
				}
			}
		}(wp.GetWorker(i), wp.Done)
	}

	wp.logger.NoticeCtx("All workers started successfully", nil)
	wp.logger.NoticeCtx(fmt.Sprintf("WorkerPool started with %d workers", wp.workerLimit), nil)
	wp.logger.NoticeCtx(fmt.Sprintf("WorkerPool is running: %t", wp.isRunning), nil)

	// Wait for all workers to finish
	wg.Wait()

	wp.logger.NoticeCtx("All workers finished processing", nil)
	wp.isRunning = false

	return wp
}
func (wp *WorkerPool) StopWorkers() {
	wp.Lock()
	defer wp.Unlock()

	if !wp.isRunning {
		wp.logger.WarnCtx("WorkerPool is not running", nil)
		return
	}

	close(wp.stopChannel)
	close(wp.jobChannel)
	wp.isRunning = false
	wp.logger.NoticeCtx("WorkerPool stopped successfully", nil)
}

func (wp *WorkerPool) GetWorker(workerID int) t.IWorker {
	wp.RLock()
	defer wp.RUnlock()
	if _, ok := wp.workerPool[workerID]; !ok {
		wp.workerPool[workerID] = NewWorker(workerID, wp.GetStopChannel(), wp.logger)
	}
	return wp.workerPool[workerID]
}
func (wp *WorkerPool) GetWorkerPool() t.IWorkerPool {
	wp.RLock()
	defer wp.RUnlock()
	return wp
}

func (wp *WorkerPool) GetStopChannel() chan struct{} {
	wp.RLock()
	defer wp.RUnlock()
	return wp.stopChannel
}
func (wp *WorkerPool) SetStopChannel(stopChannel chan struct{}) {
	wp.Lock()
	defer wp.Unlock()
	if stopChannel == nil {
		wp.logger.ErrorCtx("stop channel is nil", nil)
		return
	}
	wp.stopChannel = stopChannel
}

func (wp *WorkerPool) GetCancelChannel() chan struct{} {
	wp.RLock()
	defer wp.RUnlock()
	return wp.cancelChannel
}
func (wp *WorkerPool) SetCancelChannel(cancelChannel chan struct{}) {
	wp.Lock()
	defer wp.Unlock()
	if cancelChannel == nil {
		wp.logger.ErrorCtx("cancel channel is nil", nil)
		return
	}
	wp.cancelChannel = cancelChannel
}

func (wp *WorkerPool) GetMonitorChannel() chan t.MonitorMessage {
	wp.RLock()
	defer wp.RUnlock()
	return wp.monitorChannel
}
func (wp *WorkerPool) SetMonitorChannel(monitorChannel chan t.MonitorMessage) {
	wp.Lock()
	defer wp.Unlock()
	if monitorChannel == nil {
		wp.logger.ErrorCtx("monitor channel is nil", nil)
		return
	}
	wp.monitorChannel = monitorChannel
}

func (wp *WorkerPool) SubmitJob(job t.IJob) {
	wp.Lock()
	defer wp.Unlock()

	if wp.jobChannel == nil {
		wp.logger.ErrorCtx("job channel is not initialized", nil)
		return
	}

	select {
	case wp.jobChannel <- job:
		wp.logger.NoticeCtx(fmt.Sprintf("Job %s submitted successfully", job.GetID()), nil)
	default:
		wp.logger.WarnCtx(fmt.Sprintf("Job %s could not be submitted, worker limit reached", job.GetID()), nil)
	}
}
func (wp *WorkerPool) GetJobChannel() chan t.IJob {
	wp.RLock()
	defer wp.RUnlock()
	return wp.jobChannel
}
func (wp *WorkerPool) SetJobChannel(jobChannel chan t.IJob) {
	wp.Lock()
	defer wp.Unlock()
	if jobChannel == nil {
		wp.logger.ErrorCtx("job channel is nil", nil)
		return
	}
	wp.jobChannel = jobChannel
}

func (wp *WorkerPool) GetResultChannel() chan t.IResult {
	wp.RLock()
	defer wp.RUnlock()
	return wp.resultChannel
}
func (wp *WorkerPool) SetResultChannel(resultChannel chan t.IResult) {
	wp.Lock()
	defer wp.Unlock()
	if resultChannel == nil {
		wp.logger.ErrorCtx("result channel is nil", nil)
		return
	}
	wp.resultChannel = resultChannel
}

func (wp *WorkerPool) GetBuffersSize() int {
	wp.RLock()
	defer wp.RUnlock()
	if wp.jobChannel == nil || wp.resultChannel == nil || wp.monitorChannel == nil {
		wp.logger.ErrorCtx("One of the channels is not initialized", nil)
		return 0
	} else {
		wp.logger.NoticeCtx("Getting buffer size", nil)
		wp.logger.NoticeCtx(fmt.Sprintf("Job channel size: %d", len(wp.jobChannel)), nil)
		wp.logger.NoticeCtx(fmt.Sprintf("Result channel size: %d", len(wp.resultChannel)), nil)
		wp.logger.NoticeCtx(fmt.Sprintf("Monitor channel size: %d", len(wp.monitorChannel)), nil)
	}
	return len(wp.jobChannel) + len(wp.resultChannel) + len(wp.monitorChannel)
}
func (wp *WorkerPool) AdjustBufferSize(autoSize bool, size int) {
	wp.Lock()
	defer wp.Unlock()

	if autoSize {
		wp.logger.NoticeCtx("Auto-sizing buffer size", nil)
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
		wp.logger.NoticeCtx("Setting buffer size manually", nil)
		wp.jobChannel = make(chan t.IJob, size)
		wp.resultChannel = make(chan t.IResult, size)
		wp.monitorChannel = make(chan t.MonitorMessage, size)
	}
}
func (wp *WorkerPool) GetLogger() l.Logger {
	wp.RLock()
	defer wp.RUnlock()
	return wp.logger
}
func (wp *WorkerPool) GetWorkerCount() int {
	wp.RLock()
	defer wp.RUnlock()
	return len(wp.workerPool)
}
func (wp *WorkerPool) GetWorkerLimit() int {
	wp.RLock()
	defer wp.RUnlock()
	return wp.workerLimit
}
func (wp *WorkerPool) SetWorkerLimit(workerLimit int) error {
	wp.Lock()
	defer wp.Unlock()
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

func (wp *WorkerPool) IsRunning() bool {
	wp.RLock()
	defer wp.RUnlock()
	return wp.isRunning
}
