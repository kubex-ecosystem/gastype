package manager

import (
	"fmt"
	"sync"

	t "github.com/faelmori/gastype/types"
	l "github.com/faelmori/logz"
)

// WorkerManager manages the pool of workers for parallel task execution
type WorkerManager struct {
	WorkerCount int
	JobQueue    chan t.IAction
	Results     chan t.IResult
	StopChannel chan struct{}
	mu          sync.Mutex
}

// NewWorkerManager creates a new instance of WorkerManager
func NewWorkerManager(workerCount int) t.IWorker {
	return &WorkerManager{
		WorkerCount: workerCount,
		JobQueue:    make(chan t.IAction, workerCount),
		Results:     make(chan t.IResult, workerCount),
		StopChannel: make(chan struct{}),
	}
}

// StartWorkers starts the pool of workers
func (wm *WorkerManager) StartWorkers() {
	var wg sync.WaitGroup

	l.Info(fmt.Sprintf("Starting workers %d", wm.WorkerCount), map[string]interface{}{"worker_count": wm.WorkerCount})
	for i := 0; i < wm.WorkerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			l.Info("Worker started", map[string]interface{}{"worker_id": workerID})

			for {
				select {
				case <-wm.StopChannel:
					l.Info("Worker stopped", map[string]interface{}{
						"worker_id": workerID,
					})
					return
				case job, ok := <-wm.JobQueue:
					if !ok {
						return
					}
					if err := job.Execute(); err != nil {
						l.Error("Error executing job", map[string]interface{}{
							"error": err.Error(),
							"job":   job.GetType(),
						})
					}
				}
			}
		}(i)
	}

	l.Info("All workers started", map[string]interface{}{"worker_count": wm.WorkerCount})
	wg.Wait()

	l.Info("All workers finished", map[string]interface{}{"worker_count": wm.WorkerCount})
}

// StopWorkers stops all workers gracefully
func (wm *WorkerManager) StopWorkers() {
	l.Info("Stopping workers", map[string]interface{}{"worker_count": wm.WorkerCount})
	close(wm.StopChannel)
	close(wm.JobQueue)
}

// GetJobQueue returns the job queue channel
func (wm *WorkerManager) GetJobQueue() chan t.IAction {
	// Ensure the JobQueue is initialized
	if wm.JobQueue == nil {
		l.Error("JobQueue is not initialized", nil)
		return nil
	}
	return wm.JobQueue
}
