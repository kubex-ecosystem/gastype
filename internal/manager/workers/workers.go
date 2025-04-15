package workers

import (
	"fmt"
	g "github.com/faelmori/gastype/internal/globals"
	t "github.com/faelmori/gastype/types"
	c "github.com/faelmori/kubex-interfaces/config"
	l "github.com/faelmori/logz"
	"reflect"
	"sync"
	"time"
)

var (
	monitorMap     = sync.Map{}
	arrChanMonitor = sync.Map{} //map[int]chan MonitorMessage
	arrResults     []c.Metadata
)

type Worker struct {
	t.IWorker
	*g.Threading

	workerID    int
	JobQueue    chan t.IAction
	Results     chan t.IResult
	StopChannel chan struct{}
	logger      l.Logger
}

// NewWorker creates a new instance
func NewWorker(workerID int, stopChannel chan struct{}, logger l.Logger) t.IWorker {
	if stopChannel == nil {
		stopChannel = make(chan struct{})
	}
	if logger == nil {
		logger = l.GetLogger("GasType")
	}
	return &Worker{
		Threading:   g.NewThreading(),
		workerID:    workerID,
		JobQueue:    make(chan t.IAction, 50),
		Results:     make(chan t.IResult, 50),
		StopChannel: stopChannel,
		logger:      logger,
	}
}

// GetID returns the worker ID
func (wm *Worker) GetID() int {
	if wm.workerID == 0 {
		wm.logger.ErrorCtx("Worker ID is not set", nil)
		return -1
	}
	return wm.workerID
}

// GetStopChannel returns the stop channel
func (wm *Worker) GetStopChannel() chan struct{} {
	if wm.StopChannel == nil {
		wm.logger.ErrorCtx("Stop channel is not initialized", nil)
		return nil
	}
	return wm.StopChannel
}

// GetJobQueue Count returns the worker count
func (wm *Worker) GetJobQueue() chan t.IAction {
	// Ensure the JobQueue is initialized
	if wm.JobQueue == nil {
		wm.logger.ErrorCtx("JobQueue is not initialized", nil)
		return nil
	}
	return wm.JobQueue
}

// SetJobQueue sets the job queue
func (wm *Worker) SetJobQueue(jobQueue chan t.IAction) { wm.JobQueue = jobQueue }

// StartWorker starts the worker threads
func (wm *Worker) StartWorker() t.IWorker {
	wm.logger.DebugCtx("Starting worker", nil)

	wm.Add(1)
	go wm.workerLoop()

	go func() {
		wm.Wait()
		close(wm.StopChannel)
	}()

	wm.logger.DebugCtx("Workers started", nil)
	return wm
}

// GetLogger returns the logger
func (wm *Worker) GetLogger() l.Logger {
	if wm.logger == nil {
		wm.logger = l.GetLogger("GasType")
	}
	return wm.logger
}

// SetLogger sets the logger
func (wm *Worker) SetLogger(logger l.Logger) {
	if logger == nil {
		wm.logger = l.GetLogger("GasType")
	} else {
		wm.logger = logger
	}
}

// workerLoop processes jobs
func (wm *Worker) workerLoop() {
	defer wm.Done()
	for {
		select {
		case <-wm.StopChannel:
			wm.logger.InfoCtx(fmt.Sprintf("Worker %d stopping", wm.workerID), nil)
			return
		case job := <-wm.JobQueue:
			wm.processJob(job)
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// processJob executes the job
func (wm *Worker) processJob(job t.IAction) {
	wm.logger.DebugCtx(fmt.Sprintf("Worker %d executing job", wm.workerID), map[string]interface{}{
		"job": job.GetType(),
	})
	if err := job.Execute(); err != nil {
		wm.logger.ErrorCtx(fmt.Sprintf("Worker %d error executing job: %v", wm.workerID, err), nil)
		return
	}
	// Send results
	for _, result := range job.GetResults() {
		wm.Results <- result
	}
	wm.logger.DebugCtx(fmt.Sprintf("Worker %d finished job", wm.workerID), nil)
}

// doEachLoop initializes the worker
func (wm *Worker) doEachLoop(i int, monPos *t.MonitorMessage) {
	if m, ok := monitorMap.Load(i); ok {
		monPos = m.(*t.MonitorMessage)
		wm.logger.WarnCtx(fmt.Sprintf("Worker %d already exists", i), nil)
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
func (wm *Worker) monitorRoutine(workerID int, chanMonitor chan t.MonitorMessage, closeChan chan struct{}) {
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
							wm.logger.WarnCtx(fmt.Sprintf("Worker %d already stopped", workerID), nil)
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
			wm.logger.NoticeCtx(fmt.Sprintf("Worker %d stopping", workerID), nil)
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
			go func(wm t.IWorker, result t.IResult) {
				// Check if the result is of type IResult

				if arrResults == nil {
					arrResults = make([]c.Metadata, 0)
				}

				wm.GetLogger().DebugCtx(fmt.Sprintf("Worker %d processing result: %v", workerID, result), nil)

				if iResult, ok := result.(t.IResult); ok {
					resultMap := iResult.ToMap()
					if resultMap == nil {
						wm.GetLogger().ErrorCtx(fmt.Sprintf("Worker %d error converting result to map", workerID), nil)
						return
					}
					arrResults = append(arrResults, resultMap)
					wm.GetLogger().DebugCtx(fmt.Sprintf("Worker %d result added to array: %v", workerID, resultMap), nil)
				} else {
					wm.GetLogger().ErrorCtx(fmt.Sprintf("Worker %d error casting result to IResult: %v", workerID, result), nil)
				}
			}(wm, result)
		default:
			// Do nothing
		}

		// Sleep for a short duration to avoid busy waiting
		time.Sleep(100 * time.Millisecond)
	}
}
