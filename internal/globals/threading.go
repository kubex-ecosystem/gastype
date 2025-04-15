package globals

import (
	t "github.com/faelmori/gastype/types"
	"sync"
)

// Threading is a struct that holds the mutexes for the spider
type Threading struct {
	// IThreading interface for threading
	t.IThreading
	// Mutex for thread safety
	mu sync.RWMutex
	// SyncGroup for synchronization
	wg sync.WaitGroup
}

// NewThreading creates a new Threading instance
func NewThreading() *Threading {
	return &Threading{
		mu: sync.RWMutex{},
		wg: sync.WaitGroup{},
	}
}

// TryLock tries to lock the mutex
func (k *Threading) TryLock() bool { return k.mu.TryLock() }

// Lock locks the mutex
func (k *Threading) Lock() {
	k.mu.Lock()
}

// Unlock unlocks the mutex
func (k *Threading) Unlock() {
	k.mu.Unlock()
}

// RLock locks the mutex for reading
func (k *Threading) RLock() {
	k.mu.RLock()
}

// RUnlock unlocks the mutex for reading
func (k *Threading) RUnlock() {
	k.mu.RUnlock()
}

// LockFunc locks the mutex and executes the function
func (k *Threading) LockFunc(f func()) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if f == nil {
		return
	}
	f()
}

// UnlockFunc unlocks the mutex and executes the function
func (k *Threading) UnlockFunc(f func()) {
	k.mu.Unlock()
	defer k.mu.Lock()
	if f == nil {
		return
	}
	f()
}

// LockFuncWithArgs locks the mutex and executes the function with args
func (k *Threading) LockFuncWithArgs(f func(interface{}), args interface{}) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if f == nil {
		return
	}
	f(args)
}

// UnlockFuncWithArgs unlocks the mutex and executes the function with args
func (k *Threading) UnlockFuncWithArgs(f func(interface{}), args interface{}) {
	k.mu.Unlock()
	defer k.mu.Lock()
	if f == nil {
		return
	}
	f(args)
}

// Defer returns a function that will be executed when the mutex is unlocked
func (k *Threading) Defer() func(f func(), args interface{}) func() error {
	return func(f func(), args interface{}) func() error {
		return func() error {
			k.mu.Unlock()
			if f != nil {
				f()
			}
			return nil
		}
	}
}

// Add adds delta to the WaitGroup counter
func (k *Threading) Add(delta int) { k.wg.Add(delta) }

// Wait waits for the WaitGroup counter to reach zero
func (k *Threading) Wait() {
	k.wg.Wait()
}

// Done decrements the WaitGroup counter
func (k *Threading) Done() {
	k.wg.Done()
}

// WaitGroup returns the WaitGroup
func (k *Threading) WaitGroup() *sync.WaitGroup {
	return &k.wg
}

// WaitGroupAdd adds delta to the WaitGroup counter
func (k *Threading) WaitGroupAdd(delta int) {
	k.wg.Add(delta)
}

// WaitGroupDone decrements the WaitGroup counter
func (k *Threading) WaitGroupDone() {
	k.wg.Done()
}
