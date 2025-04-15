package types

import (
	l "github.com/faelmori/logz"
)

// IWorker defines the interface for a worker that processes jobs and interacts with a worker pool.
type IWorker interface {
	// GetID returns the unique identifier of the worker.
	GetID() int

	// GetStopChannel returns the channel used to signal the worker to stop.
	GetStopChannel() chan struct{}

	// GetJobQueue returns the channel where jobs are submitted to the worker.
	GetJobQueue() chan IAction

	// SetJobQueue sets the channel where jobs are submitted to the worker.
	SetJobQueue(chan IAction)

	// StartWorker initializes and starts the worker.
	StartWorker() IWorker

	// GetLogger returns the logger instance used by the worker.
	GetLogger() l.Logger

	// SetLogger sets the logger instance for the worker.
	SetLogger(l.Logger)

	// workerLoop is the main loop of the worker, processing jobs and handling signals.
	workerLoop()

	// processJob processes a single job submitted to the worker.
	processJob(IAction)

	// doEachLoop performs actions during each iteration of the worker loop.
	doEachLoop(int, *MonitorMessage)

	// monitorRoutine handles monitoring of the worker's state and activity.
	monitorRoutine(int, chan MonitorMessage, chan struct{})
}

// IWorkerPool defines the interface for a pool of workers that manage job processing.
type IWorkerPool interface {
	// StartWorkers initializes and starts all workers in the pool.
	StartWorkers() IWorkerPool

	// StopWorkers stops all workers in the pool.
	StopWorkers()

	// GetWorker retrieves a worker by its index.
	GetWorker(int) IWorker

	// GetWorkerPool returns the worker pool instance.
	GetWorkerPool() IWorkerPool

	// GetStopChannel returns the channel used to signal the pool to stop.
	GetStopChannel() chan struct{}

	// SetStopChannel sets the channel used to signal the pool to stop.
	SetStopChannel(chan struct{})

	// GetCancelChannel returns the channel used to cancel operations in the pool.
	GetCancelChannel() chan struct{}

	// SetCancelChannel sets the channel used to cancel operations in the pool.
	SetCancelChannel(chan struct{})

	// GetMonitorChannel returns the channel used for monitoring messages.
	GetMonitorChannel() chan MonitorMessage

	// SetMonitorChannel sets the channel used for monitoring messages.
	SetMonitorChannel(chan MonitorMessage)

	// SubmitJob submits a job to the pool for processing.
	SubmitJob(IJob)

	// GetJobChannel returns the channel where jobs are submitted to the pool.
	GetJobChannel() chan IJob

	// SetJobChannel sets the channel where jobs are submitted to the pool.
	SetJobChannel(chan IJob)

	// GetResultChannel returns the channel where job results are sent.
	GetResultChannel() chan IResult

	// SetResultChannel sets the channel where job results are sent.
	SetResultChannel(chan IResult)

	// GetBuffersSize returns the size of the buffers used in the pool.
	GetBuffersSize() int

	// AdjustBufferSize adjusts the size of the buffers in the pool.
	AdjustBufferSize(bool, int)

	// GetLogger returns the logger instance used by the pool.
	GetLogger() l.Logger

	// GetWorkerCount returns the current number of workers in the pool.
	GetWorkerCount() int

	// GetWorkerLimit returns the maximum number of workers allowed in the pool.
	GetWorkerLimit() int

	// SetWorkerLimit sets the maximum number of workers allowed in the pool.
	SetWorkerLimit(int) error

	// IsRunning checks if the worker pool is currently running.
	IsRunning() bool
}

// IWorkerManager defines the interface for managing a worker pool and its operations.
type IWorkerManager interface {
	// StartWorkers initializes and starts all workers managed by the manager.
	StartWorkers()

	// StopWorkers stops all workers managed by the manager.
	StopWorkers()

	// GetWorkerPool returns the worker pool instance managed by the manager.
	GetWorkerPool() IWorkerPool

	// SetWorkerPool sets the worker pool instance to be managed by the manager.
	SetWorkerPool(IWorkerPool)

	// GetLogger returns the logger instance used by the manager.
	GetLogger() l.Logger

	// SetLogger sets the logger instance for the manager.
	SetLogger(l.Logger)

	// GetStopChannel returns the channel used to signal the manager to stop.
	GetStopChannel() chan struct{}

	// SetStopChannel sets the channel used to signal the manager to stop.
	SetStopChannel(chan struct{})

	// GetJobChannel returns the channel where jobs are submitted to the manager.
	GetJobChannel() chan IAction

	// SetJobChannel sets the channel where jobs are submitted to the manager.
	SetJobChannel(chan IAction)

	// GetCancelChannel returns the channel used to cancel operations in the manager.
	GetCancelChannel() chan struct{}

	// SetCancelChannel sets the channel used to cancel operations in the manager.
	SetCancelChannel(chan struct{})

	// GetMonitorChannel returns the channel used for monitoring messages.
	GetMonitorChannel() chan MonitorMessage

	// SetMonitorChannel sets the channel used for monitoring messages.
	SetMonitorChannel(chan MonitorMessage)

	// GetWorkerCount returns the current number of workers managed by the manager.
	GetWorkerCount() int

	// GetWorkerLimit returns the maximum number of workers allowed by the manager.
	GetWorkerLimit() int

	// GetBuffersSize returns the size of the buffers used by the manager.
	GetBuffersSize() int

	// AdjustBufferSize adjusts the size of the buffers used by the manager.
	AdjustBufferSize(bool, int)

	// SetWorkerLimit sets the maximum number of workers allowed by the manager.
	SetWorkerLimit(int) error

	// SetWorkerCount sets the current number of workers managed by the manager.
	SetWorkerCount(int) error
}
