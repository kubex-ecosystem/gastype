package manager

import (
	"fmt"
	"sync"

	t "github.com/kubex-ecosystem/gastype/interfaces"

	gl "github.com/kubex-ecosystem/logz/logger"
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
			gl.Log("info", fmt.Sprintf("Worker started: %d", workerID))

			for {
				select {
				case <-wm.StopChannel:
					gl.Log("info", fmt.Sprintf("Worker stopped: %d", workerID))
					return
				case job, ok := <-wm.JobQueue:
					if !ok {
						return
					}
					if err := job.Execute(); err != nil {
						gl.Log("error", fmt.Sprintf("Error executing job: %s, job: %s", err.Error(), job.GetType()))
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
