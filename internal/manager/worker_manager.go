package manager

import (
	g "github.com/faelmori/gastype/internal/globals"
	t "github.com/faelmori/gastype/types"
	kc "github.com/faelmori/kubex-interfaces/config"
	tl "github.com/faelmori/kubex-interfaces/tools"
	"github.com/google/uuid"
)

// WorkerManager with corrections
type WorkerManager struct {
	// IWorkerManager interface for worker management
	t.IWorkerManager
	// Threading interface for threading
	// Mutex for thread safety
	// SyncGroup for synchronization
	*g.Threading
	// ID of the worker manager
	ID string
	// Properties map for storing properties
	Properties map[string]kc.Property[any]
}

// NewWorkerManager creates a new instance of WorkerManager
func NewWorkerManager(workerLimit int) *WorkerManager {
	wp := &WorkerManager{
		ID:         uuid.NewString(),
		Threading:  g.NewThreading(),
		Properties: make(map[string]kc.Property[any]),
	}

	wp.Properties["workerStatus"] = kc.NewProperty[any]("workerStatus", make(map[int]*t.MonitorMessage))
	wp.Properties["workerStatus"].SetMetadata("description", "Status of each worker")
	wp.Properties["workerStatus"].SetMetadata("type", "map")
	wp.Properties["workerStatus"].SetDefaultValue(make(map[int]*t.MonitorMessage))

	wp.Properties["workerLimit"] = kc.NewProperty[any]("workerLimit", workerLimit)
	wp.Properties["workerCount"] = kc.NewProperty[any]("workerCount", 0)
	wp.Properties["workerPool"] = kc.NewProperty[any]("workerPool", make(map[int]t.IWorker))
	wp.Properties["workerPool"].SetChannel(tl.NewChannel[t.MonitorMessage, int]("workerPool", &t.MonitorMessage{}, 20))

	return wp
}
