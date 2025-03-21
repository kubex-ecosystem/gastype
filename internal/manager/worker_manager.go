package manager

import (
	"sync"

	l "github.com/faelmori/gastype/log"
	t "github.com/faelmori/gastype/types"
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

	for i := 0; i < wm.WorkerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			l.Info("Worker started", map[string]interface{}{
				"worker_id": workerID,
			})

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

	wg.Wait()
}

// StopWorkers stops all workers gracefully
func (wm *WorkerManager) StopWorkers() {
	close(wm.StopChannel)
	close(wm.JobQueue)
}

// GetJobQueue returns the job queue channel
func (wm *WorkerManager) GetJobQueue() chan t.IAction {
	return wm.JobQueue
}
