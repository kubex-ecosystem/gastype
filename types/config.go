package types

type IConfig interface {
	Load() error
	GetDir() string
	GetWorkerCount() int
	GetWorkerLimit() int
	GetOutputFile() string

	SetDir(string)
	SetWorkerLimit(int)
	SetOutputFile(string)
}
