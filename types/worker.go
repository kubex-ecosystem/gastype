package types

import (
	l "github.com/faelmori/logz"
)

type IWorker interface {
	StartWorkers() IWorker
	StopWorkers()
	GetJobQueue() chan IAction
}

type IWorkerPool interface {
	StartWorkers() IWorkerPool

	SubmitJob(job IJob)
	StopWorkers()

	GetBuffersSize() int
	AjustBufferSize(autoSize bool, size int)
	GetStopChannel() chan struct{}
	GetLogger() l.Logger
	GetMonitorChannel() chan MonitorMessage

	GetJobChannel() chan IJob
	GetResultChannel() chan IResult

	GetWorkerCount() int
	GetWorkerLimit() int
	GetWorkerPool() map[int]IWorker

	SetWorkerLimit(workerLimit int) error

	IsRunning() bool
}

type IJob interface {
	IAction
	GetID() string
}
