package types

import l "github.com/faelmori/logz"

type IConfig interface {
	Load() error
	GetDir() string
	GetWorkerCount() int
	GetWorkerLimit() int
	GetOutputFile() string

	SetDir(string)
	SetWorkerLimit(int)
	SetOutputFile(string)

	GetLogger() l.Logger
	SetLogger(l.Logger)
	GetChanCtl() chan interface{}
}
