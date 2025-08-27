package interfaces

type IWorker interface {
	StartWorkers()
	StopWorkers()
	GetJobQueue() chan IAction
}

type IWorkerPool interface {
	SubmitJob(job IJob)
	StopWorkers()

	GetJobChannel() chan IJob
	GetResultChannel() chan IResult

	GetWorkerCount() int
	GetWorkerLimit() int
	GetWorkerPool() []IWorker

	SetWorkerLimit(workerLimit int)

	IsRunning() bool
}

type IJob interface {
	IAction
	GetID() string
}
