package types

type IConfig interface {
	Load() error
	GetDir() string
	GetWorkerCount() int
	GetOutputFile() string

	SetDir(string)
	SetWorkerLimit(int)
	SetOutputFile(string)
}
