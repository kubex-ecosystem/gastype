package manager

import (
	"fmt"
	"github.com/faelmori/gastype/internal/manager/workers"
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

// WorkerManager manages the pool of workers for parallel task execution
type WorkerManager struct {
	WorkerCount int
	JobQueue    chan t.IAction
	Results     chan t.IResult
	StopChannel chan struct{}
	logger      l.Logger
	mu          sync.Mutex
}

// NewWorkerManager creates a new instance of WorkerManager
func NewWorkerManager(workerCount int, logger l.Logger) t.IWorker {
	if arrResults == nil {
		arrResults = make([]c.Metadata, 0)
	}
	if logger == nil {
		logger = l.GetLogger("GasType")
	}
	if workerCount <= 1 {
		workerCount = 2
	}

	workerPool := workers.NewWorkerPool(workerCount, logger)
	if workerPool == nil {
		logger.ErrorCtx("Error creating worker pool", nil)
		return nil
	}

	return &WorkerManager{
		WorkerCount: workerCount,
		JobQueue:    make(chan t.IAction, 50),
		Results:     make(chan t.IResult, 50),
		StopChannel: make(chan struct{}, 5),
		logger:      logger,
		mu:          sync.Mutex{},
	}
}

func (wm *WorkerManager) doEachLoop(i int, monPos *t.MonitorMessage) {
	if m, ok := monitorMap.Load(i); ok {
		monPos = m.(*t.MonitorMessage)
		wm.logger.DebugCtx(fmt.Sprintf("Worker %d already exists", i), nil)
		return
	} else {
		monitorMap.Store(i, monPos)
		ch := make(chan t.MonitorMessage, 10)
		arrChanMonitor.Store(i, ch)
		go wm.monitorRoutine(i, ch, wm.StopChannel)
		ch <- *monPos
	}
}

func (wm *WorkerManager) monitorRoutine(workerID int, chanMonitor chan t.MonitorMessage, closeChan chan struct{}) {
	// Monitor the worker's channel
	for {
		select {
		case msg := <-chanMonitor:
			wm.logger.DebugCtx(fmt.Sprintf("Worker %d received message: %v", workerID, msg), nil)
			if reflect.TypeOf(msg) == reflect.TypeOf(t.MonitorMessage{}) {
				old, ok := monitorMap.Load(workerID)
				if ok {
					if oldMsg, ok := old.(*t.MonitorMessage); ok {
						if oldMsg.Status == "stopped" {
							wm.logger.DebugCtx(fmt.Sprintf("Worker %d already stopped", workerID), nil)
							return
						}
					} else {
						wm.logger.ErrorCtx(fmt.Sprintf("Worker %d message type mismatch", workerID), nil)
						return
					}
				}
				monitorMap.CompareAndSwap(workerID, old, msg)
				wm.logger.DebugCtx(fmt.Sprintf("Worker %d message updated: %v", workerID, msg), nil)
			}
		case <-wm.StopChannel:
			wm.logger.DebugCtx(fmt.Sprintf("Worker %d stopping", workerID), nil)
			// Print all worker information
			monitorMap.Range(func(key, value interface{}) bool {
				wm.logger.DebugCtx(fmt.Sprintf("Worker %d: Status=%s, JobType=%s", key, value.(*t.MonitorMessage).Status, value.(*t.MonitorMessage).JobType), nil)
				return true
			})
			wm.logger.DebugCtx(fmt.Sprintf("Worker %d stopping", workerID), nil)
			closeChan <- struct{}{}
			return
		case result := <-wm.Results:
			wm.logger.DebugCtx(fmt.Sprintf("Worker %d received result: %v", workerID, result), nil)
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
					wm.logger.DebugCtx(fmt.Sprintf("Worker %d result added to array: %v", workerID, resultMap), nil)
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

// StartWorkers starts the pool of workers
func (wm *WorkerManager) StartWorkers() t.IWorker {
	if arrResults == nil {
		arrResults = make([]c.Metadata, 0)
	}

	// Launch worker goroutines without waiting for them to finish
	for i := 0; i < wm.WorkerCount; i++ {
		wm.logger.DebugCtx(fmt.Sprintf("Starting worker %d", i), nil)
		go func(workerID int) {
			// Ensure monitor channel setup
			chM, chanMonitorOk := arrChanMonitor.Load(workerID)
			var chanMonitor chan t.MonitorMessage
			if !chanMonitorOk {
				chanMonitor = make(chan t.MonitorMessage, 10)
				arrChanMonitor.Store(workerID, chanMonitor)
				monPos := &t.MonitorMessage{WorkerID: workerID, Status: "starting", JobType: "none", StartTime: time.Now()}
				wm.doEachLoop(workerID, monPos)
			} else {
				if ch, ok := chM.(chan t.MonitorMessage); ok {
					chanMonitor = ch
				}
			}
			wm.logger.DebugCtx(fmt.Sprintf("Worker %d started", workerID), nil)
			for {
				select {
				case <-wm.StopChannel:
					wm.logger.DebugCtx(fmt.Sprintf("Worker %d stopping", workerID), nil)
					return
				case job, ok := <-wm.JobQueue:
					if !ok {
						wm.logger.WarnCtx(fmt.Sprintf("Worker %d received nil job", workerID), nil)
						return
					}
					wm.logger.InfoCtx(fmt.Sprintf("Worker %d received job: %v", workerID, job), nil)
					if err := job.Execute(); err != nil {
						wm.logger.ErrorCtx(fmt.Sprintf("Worker %d error executing job: %s", workerID, err.Error()), nil)
					} else {
						wm.logger.DebugCtx(fmt.Sprintf("Worker %d finished executing job", workerID), nil)
						if job.IsRunning() {
							for _, result := range job.GetResults() {
								if result != nil {
									wm.Results <- result.(t.IResult)
								} else {
									wm.logger.ErrorCtx(fmt.Sprintf("Worker %d job result is nil", workerID), nil)
								}
							}
						}
					}
				default:
					time.Sleep(100 * time.Millisecond)
				}
			}
		}(i)
	}
	wm.logger.DebugCtx("All workers launched", nil)
	// Immediately return without waiting.
	return wm
}

// StopWorkers stops all workers gracefully
func (wm *WorkerManager) StopWorkers() {
	l.DebugCtx("Stopping workers", map[string]interface{}{"worker_count": wm.WorkerCount})
	close(wm.StopChannel)
	close(wm.JobQueue)
}

// GetJobQueue returns the job queue channel
func (wm *WorkerManager) GetJobQueue() chan t.IAction {
	// Ensure the JobQueue is initialized
	if wm.JobQueue == nil {
		wm.logger.ErrorCtx("JobQueue is not initialized", nil)
		return nil
	}
	return wm.JobQueue
}
