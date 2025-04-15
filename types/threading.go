package types

import "sync"

// IDefer is a struct that holds the defer function
type IDefer interface {
	Defer() func(f func(), args interface{}) func() error
}

// IMutex is a struct that holds the mutexes
type IMutex interface {
	TryLock() bool
	Lock()
	Unlock()
	RLock()
	RUnlock()
	LockFunc(func())
	UnlockFunc(func())
	LockFuncWithArgs(func(interface{}), interface{})
	UnlockFuncWithArgs(func(interface{}), interface{})
}

// ISync is a struct that holds the sync.WaitGroup
type ISync interface {
	Add(delta int)
	Wait()
	Done()
	WaitGroup() *sync.WaitGroup
	WaitGroupAdd(delta int)
	WaitGroupDone()
}

// IThreading is a struct that holds the defer function
type IThreading interface {
	IMutex
	ISync
	IDefer
}
