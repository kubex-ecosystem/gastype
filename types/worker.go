package types

type IWorker interface {
	StartWorkers()
	StopWorkers()
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
	GetID() string
	GetType() string
	GetData() map[string]interface{}
	GetResults() map[string]interface{}
	GetStatus() string
	GetErrors() []error
}
