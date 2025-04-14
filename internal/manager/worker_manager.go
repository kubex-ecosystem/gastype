package manager

import (
	"fmt"
	c "github.com/faelmori/kubex-interfaces/config"
	"reflect"
	"sync"
	"time"

	t "github.com/faelmori/gastype/types"
	l "github.com/faelmori/logz"
)

var (
	monitorMap     = sync.Map{}
	arrChanMonitor = sync.Map{} //map[int]chan MonitorMessage
	arrResults     []c.Metadata
)

// WorkerManager with corrections
type WorkerManager struct {
	mu          sync.Mutex
	WorkerPool  t.IWorkerPool
	WorkerCount int
	JobQueue    chan t.IAction
	Results     chan t.IResult
	StopChannel chan struct{}
	logger      l.Logger
}

// NewWorkerManager creates a new instance
func NewWorkerManager(workerCount int, workerPool t.IWorkerPool, logger l.Logger) t.IWorker {
	if workerCount <= 0 {
		workerCount = 1
	}
	return &WorkerManager{
		mu:          sync.Mutex{},
		WorkerCount: workerCount,
		WorkerPool:  workerPool,
		JobQueue:    make(chan t.IAction, 50),
		Results:     make(chan t.IResult, 50),
		StopChannel: make(chan struct{}),
		logger:      logger,
	}
}

func (wm *WorkerManager) StartWorkers() t.IWorker {
	var wg sync.WaitGroup
	wm.logger.InfoCtx("Starting workers", map[string]interface{}{"worker_count": wm.WorkerCount})

	for i := 0; i < wm.WorkerCount; i++ {
		wg.Add(1)
		go wm.workerLoop(i, &wg)
	}
	wg.Wait()

	wm.logger.InfoCtx("All workers started", nil)

	return wm
}
func (wm *WorkerManager) GetWorkerPool() t.IWorkerPool {
	if wm.WorkerPool == nil {
		wm.logger.ErrorCtx("Worker pool not initialized", nil)
		return nil
	}
	return wm.WorkerPool
}
func (wm *WorkerManager) StopWorkers() {
	wm.logger.InfoCtx("Stopping all workers", nil)
	close(wm.StopChannel)
	close(wm.JobQueue)
}
func (wm *WorkerManager) GetJobQueue() chan t.IAction {
	// Ensure the JobQueue is initialized
	if wm.JobQueue == nil {
		wm.logger.ErrorCtx("JobQueue is not initialized", nil)
		return nil
	}
	return wm.JobQueue
}
func (wm *WorkerManager) SetJobQueue(jobQueue chan t.IAction) { wm.JobQueue = jobQueue }

// workerLoop processes jobs
func (wm *WorkerManager) workerLoop(workerID int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-wm.StopChannel:
			wm.logger.InfoCtx(fmt.Sprintf("Worker %d stopping", workerID), nil)
			return
		case job := <-wm.JobQueue:
			wm.processJob(workerID, job)
		default:
			// Avoid busy waiting
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// processJob executes the job
func (wm *WorkerManager) processJob(workerID int, job t.IAction) {
	wm.logger.InfoCtx(fmt.Sprintf("Worker %d executing job", workerID), map[string]interface{}{
		"job": job.GetType(),
	})
	if err := job.Execute(); err != nil {
		wm.logger.ErrorCtx(fmt.Sprintf("Worker %d error executing job: %v", workerID, err), nil)
		return
	}
	// Send results
	for _, result := range job.GetResults() {
		wm.Results <- result
	}
	wm.logger.InfoCtx(fmt.Sprintf("Worker %d finished job", workerID), nil)
}

// doEachLoop initializes the worker
func (wm *WorkerManager) doEachLoop(i int, monPos *t.MonitorMessage) {
	if m, ok := monitorMap.Load(i); ok {
		monPos = m.(*t.MonitorMessage)
		wm.logger.NoticeCtx(fmt.Sprintf("Worker %d already exists", i), nil)
		return
	} else {
		monitorMap.Store(i, monPos)
		ch := make(chan t.MonitorMessage, 10)
		arrChanMonitor.Store(i, ch)
		go wm.monitorRoutine(i, ch, wm.StopChannel)
		ch <- *monPos
	}
}

// monitorRoutine monitors the worker's channel
func (wm *WorkerManager) monitorRoutine(workerID int, chanMonitor chan t.MonitorMessage, closeChan chan struct{}) {
	// Monitor the worker's channel
	for {
		select {
		case msg := <-chanMonitor:
			wm.logger.NoticeCtx(fmt.Sprintf("Worker %d received message: %v", workerID, msg), nil)
			if reflect.TypeOf(msg) == reflect.TypeOf(t.MonitorMessage{}) {
				old, ok := monitorMap.Load(workerID)
				if ok {
					if oldMsg, ok := old.(*t.MonitorMessage); ok {
						if oldMsg.Status == "stopped" {
							wm.logger.NoticeCtx(fmt.Sprintf("Worker %d already stopped", workerID), nil)
							return
						}
					} else {
						wm.logger.ErrorCtx(fmt.Sprintf("Worker %d message type mismatch", workerID), nil)
						return
					}
				}
				monitorMap.CompareAndSwap(workerID, old, msg)
				wm.logger.NoticeCtx(fmt.Sprintf("Worker %d message updated: %v", workerID, msg), nil)
			}
		case <-wm.StopChannel:
			wm.logger.NoticeCtx(fmt.Sprintf("Worker %d stopping", workerID), nil)
			// Print all worker information
			monitorMap.Range(func(key, value interface{}) bool {
				wm.logger.NoticeCtx(fmt.Sprintf("Worker %d: Status=%s, JobType=%s", key, value.(*t.MonitorMessage).Status, value.(*t.MonitorMessage).JobType), nil)
				return true
			})
			wm.logger.NoticeCtx(fmt.Sprintf("Worker %d stopping", workerID), nil)
			closeChan <- struct{}{}
			return
		case result := <-wm.Results:
			wm.logger.NoticeCtx(fmt.Sprintf("Worker %d received result: %v", workerID, result), nil)
			go func(wm *WorkerManager, result t.IResult) {
				// Check if the result is of type IResult
				wm.mu.Lock()
				defer wm.mu.Unlock()
				if arrResults == nil {
					arrResults = make([]c.Metadata, 0)
				}

				wm.logger.InfoCtx(fmt.Sprintf("Worker %d processing result: %v", workerID, result), nil)

				if iResult, ok := result.(t.IResult); ok {
					resultMap := iResult.ToMap()
					if resultMap == nil {
						wm.logger.ErrorCtx(fmt.Sprintf("Worker %d error converting result to map", workerID), nil)
						return
					}
					arrResults = append(arrResults, resultMap)
					wm.logger.NoticeCtx(fmt.Sprintf("Worker %d result added to array: %v", workerID, resultMap), nil)
				} else {
					wm.logger.ErrorCtx(fmt.Sprintf("Worker %d error casting result to IResult: %v", workerID, result), nil)
				}
			}(wm, result)
		default:
			// Do nothing
		}

		// Sleep for a short duration to avoid busy waiting
		time.Sleep(100 * time.Millisecond)
	}
}
