package utils

import (
	"fmt"
	"sync"

	g "github.com/faelmori/gastype/internal/globals"
	t "github.com/faelmori/gastype/types"
	l "github.com/faelmori/logz"
)

// WorkerPool manages workers and their assigned jobs
type WorkerPool struct {
	workerCount   int
	workerLimit   int
	workerPool    []t.IWorker
	jobChannel    chan t.IJob
	resultChannel chan t.IResult
	stopChannel   chan struct{}
	mu            sync.Mutex
	isRunning     bool
}

// NewWorkerPool creates a new WorkerPool instance
func NewWorkerPool(workerLimit int) t.IWorkerPool {
	return &WorkerPool{
		workerLimit:   workerLimit,
		jobChannel:    make(chan t.IJob, workerLimit),
		resultChannel: make(chan t.IResult, workerLimit),
		stopChannel:   make(chan struct{}),
		isRunning:     false,
	}
}

// SubmitJob adds a job to the jobChannel for processing
func (wp *WorkerPool) SubmitJob(job t.IJob) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.jobChannel == nil {
		l.Error("job channel is not initialized", nil)
		return
	}

	select {
	case wp.jobChannel <- job:
		l.Info(fmt.Sprintf("Job %s submitted successfully", job.GetID()), nil)
	default:
		l.Warn(fmt.Sprintf("Job %s could not be submitted, worker limit reached", job.GetID()), nil)
	}
}

// StartWorkers initializes workers and begins processing jobs
func (wp *WorkerPool) StartWorkers() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.isRunning {
		l.Warn("WorkerPool is already running", nil)
		return
	}

	wp.isRunning = true
	var wg sync.WaitGroup

	for i := 0; i < wp.workerLimit; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			l.Info(fmt.Sprintf("Worker %d started", workerID), nil)

			for {
				select {
				case <-wp.stopChannel:
					l.Info(fmt.Sprintf("Worker %d stopped", workerID), nil)
					return
				case job, ok := <-wp.jobChannel:
					if !ok {
						l.Warn(fmt.Sprintf("Worker %d received nil job", workerID), nil)
						return
					}

					if err := job.Execute(); err != nil {
						l.Error(fmt.Sprintf("Error executing job %s: %v", job.GetID(), err), nil)
					} else {
						l.Info(fmt.Sprintf("Job %s completed successfully", job.GetID()), nil)

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

	l.Info("All workers started successfully", nil)
	l.Info(fmt.Sprintf("WorkerPool started with %d workers", wp.workerLimit), nil)
	l.Info(fmt.Sprintf("WorkerPool is running: %t", wp.isRunning), nil)

	// Wait for all workers to finish
	wg.Wait()

	l.Info("All workers finished processing", nil)
	wp.isRunning = false
}

// StopWorkers stops all workers and closes necessary channels
func (wp *WorkerPool) StopWorkers() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if !wp.isRunning {
		l.Warn("WorkerPool is not running", nil)
		return
	}

	close(wp.stopChannel)
	close(wp.jobChannel)
	wp.isRunning = false
	l.Info("WorkerPool stopped successfully", nil)
}

// Getters
func (wp *WorkerPool) GetJobChannel() chan t.IJob       { return wp.jobChannel }
func (wp *WorkerPool) GetResultChannel() chan t.IResult { return wp.resultChannel }
func (wp *WorkerPool) GetWorkerCount() int              { return len(wp.workerPool) }
func (wp *WorkerPool) GetWorkerLimit() int              { return wp.workerLimit }
func (wp *WorkerPool) GetWorkerPool() []t.IWorker       { return wp.workerPool }
func (wp *WorkerPool) IsRunning() bool                  { return wp.isRunning }

// Setters
func (wp *WorkerPool) SetWorkerLimit(workerLimit int) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	wp.workerLimit = workerLimit
}
