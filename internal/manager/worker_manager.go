package manager

import (
	"fmt"
	"reflect"
	"sync"
	"time"

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
		JobQueue:    make(chan t.IAction, 50),
		Results:     make(chan t.IResult, 50),
		StopChannel: make(chan struct{}, 5),
	}
}

type MonitorMessage struct {
	WorkerID  int
	Status    string
	JobType   string
	StartTime time.Time
	EndTime   time.Time
	Message   []string
}

var (
	monitorMap     = sync.Map{}
	arrChanMonitor map[int]chan MonitorMessage
)

func (wm *WorkerManager) doEachLoop(i int, monPos *MonitorMessage) {
	if arrChanMonitor == nil {
		arrChanMonitor = make(map[int]chan MonitorMessage)
	}

	if m, ok := monitorMap.Load(i); ok {
		monPos = m.(*MonitorMessage)

		l.Info(fmt.Sprintf("Worker %d already exists", i), nil)

		return
	} else {
		monitorMap.Store(i, monPos)
		arrChanMonitor[i] = make(chan MonitorMessage, 50)
		arrChanMonitor[i] <- *monPos

		go wm.monitorRoutine(i, arrChanMonitor[i], wm.StopChannel)
	}
}

func (wm *WorkerManager) monitorRoutine(workerID int, chanMonitor chan MonitorMessage, closeChan chan struct{}) {
	l.GetLogger("GasType")
	// Monitor the worker's channel
	for {
		select {
		case msg := <-chanMonitor:
			l.Info(fmt.Sprintf("Worker %d received message: %v", workerID, msg), nil)
			if reflect.TypeOf(msg) == reflect.TypeOf(MonitorMessage{}) {
				old, ok := monitorMap.Load(workerID)
				if ok {
					oldMsg := old.(MonitorMessage)
					if oldMsg.Status == "stopped" {
						l.Info(fmt.Sprintf("Worker %d already stopped", workerID), nil)
						return
					}
				}
				monitorMap.CompareAndSwap(workerID, old, msg)
			}
		case <-wm.StopChannel:
			l.Info(fmt.Sprintf("Worker %d stopping", workerID), nil)
			// Print all worker information
			monitorMap.Range(func(key, value interface{}) bool {
				l.Info(fmt.Sprintf("Worker %d: Status=%s, JobType=%s", key, value.(*MonitorMessage).Status, value.(*MonitorMessage).JobType), nil)
				return true
			})
			// Close the channel
			closeChan <- struct{}{}
			return
		}
	}
}

// StartWorkers starts the pool of workers
func (wm *WorkerManager) StartWorkers() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	l.Info("Starting workers", nil)
	var wg sync.WaitGroup
	// Ensure the JobQueue is initialized
	if wm.JobQueue == nil {
		l.Error("JobQueue is not initialized", nil)
		return
	}

	l.Info("Initializing JobQueue", nil)
	l.Debug("Clearing JobQueue", nil)
	// Clear the JobQueue channel
	if arrChanMonitor == nil {
		monitorMap.Clear()
	}

	l.Info(fmt.Sprintf("Starting workers %d", wm.WorkerCount), map[string]interface{}{"worker_count": wm.WorkerCount})

	for i := 0; i < wm.WorkerCount; i++ {
		l.Debug(fmt.Sprintf("Starting worker %d", i), nil)

		// Add a wait group for each worker
		wg.Add(1)

		// Start a worker goroutine
		go func(workerID int, chanMonitor chan MonitorMessage, wg *sync.WaitGroup) {
			// Defer the wait group done call
			defer wg.Done()

			monPos := &MonitorMessage{WorkerID: i, Status: "starting", JobType: "none", StartTime: time.Now(), EndTime: time.Time{}, Message: []string{}}

			wm.doEachLoop(i, monPos)

			l.Info("Worker started", map[string]interface{}{"worker_id": workerID})
			for {
				select {
				case <-wm.StopChannel:
					l.Info("Worker stopped", map[string]interface{}{
						"worker_id": workerID,
					})
					wkr, ok := monitorMap.Load(workerID)
					if ok {
						monPos := wkr.(*MonitorMessage)
						monPos.Status = "stopped"
						monPos.EndTime = time.Now()
						monPos.Message = append(monPos.Message, fmt.Sprintf("Worker %d stopped", workerID))
						chanMonitor <- *monPos
						l.Info(monPos.Message[len(monPos.Message)-1], nil)
					} else {
						l.Error(fmt.Sprintf("Worker %d not found", workerID), nil)
					}
					return
				case job, ok := <-wm.JobQueue:
					wkr, ok := monitorMap.Load(workerID)
					if ok {
						monPos := wkr.(*MonitorMessage)
						monPos.Status = "executing job"
						monPos.Message = append(monPos.Message, fmt.Sprintf("Worker %d executing job", workerID))
						chanMonitor <- *monPos
						l.Info(monPos.Message[len(monPos.Message)-1], nil)

						// Just execute the job if the worker was found in the map
						if err := job.Execute(); err != nil {
							monPos.Status = "error"
							monPos.Message = append(monPos.Message, fmt.Sprintf("Worker %d error executing job: %s", workerID, err.Error()))
							l.Error(monPos.Message[len(monPos.Message)-1], map[string]interface{}{
								"error": err.Error(),
								"job":   job.GetType(),
							})
							continue
						}
					} else {
						l.Error(fmt.Sprintf("Worker %d not found", workerID), nil)
					}
				}

				// Check if the job channel is closed
				time.Sleep(100 * time.Millisecond)
			}
		}(i, arrChanMonitor[i], &wg)
	}

	l.Info("All workers started", nil)

	// Wait for all workers to finish
	defer func() {
		defer close(wm.JobQueue)
	}()

	wg.Wait()

	l.Info("All workers finished processing", nil)

	if wm.mu.TryLock() {
		defer wm.mu.Unlock()
	} else {
		wm.mu.Unlock()
	}
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
