package types

import (
	l "github.com/faelmori/logz"
)

type IWorker interface {
	GetID() int
	GetStopChannel() chan struct{}
	GetJobQueue() chan IAction
	SetJobQueue(chan IAction)
	StartWorker() IWorker
	GetLogger() l.Logger
	SetLogger(l.Logger)
	workerLoop()
	processJob(IAction)
	doEachLoop(int, *MonitorMessage)
	monitorRoutine(int, chan MonitorMessage, chan struct{})
}

type IWorkerPool interface {
	StartWorkers() IWorkerPool
	StopWorkers()

	GetWorker(int) IWorker
	GetWorkerPool() IWorkerPool

	GetStopChannel() chan struct{}
	SetStopChannel(chan struct{})

	GetCancelChannel() chan struct{}
	SetCancelChannel(chan struct{})

	GetMonitorChannel() chan MonitorMessage
	SetMonitorChannel(chan MonitorMessage)

	SubmitJob(IJob)
	GetJobChannel() chan IJob
	SetJobChannel(chan IJob)

	GetResultChannel() chan IResult
	SetResultChannel(chan IResult)

	GetBuffersSize() int
	AdjustBufferSize(bool, int)
	GetLogger() l.Logger
	GetWorkerCount() int
	GetWorkerLimit() int
	SetWorkerLimit(int) error

	IsRunning() bool
}

type IWorkerManager interface {
	StartWorkers()
	StopWorkers()

	GetWorkerPool() IWorkerPool
	SetWorkerPool(IWorkerPool)

	GetLogger() l.Logger
	SetLogger(l.Logger)

	GetStopChannel() chan struct{}
	SetStopChannel(chan struct{})

	GetJobChannel() chan IAction
	SetJobChannel(chan IAction)

	GetCancelChannel() chan struct{}
	SetCancelChannel(chan struct{})

	GetMonitorChannel() chan MonitorMessage
	SetMonitorChannel(chan MonitorMessage)

	GetWorkerCount() int
	GetWorkerLimit() int
	GetBuffersSize() int
	AdjustBufferSize(bool, int)
	SetWorkerLimit(int) error
	SetWorkerCount(int) error
}
