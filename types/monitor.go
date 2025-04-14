package types

import "time"

type MonitorMessage struct {
	WorkerID  int
	Status    string
	JobType   string
	StartTime time.Time
	EndTime   time.Time
	Message   []string
}
