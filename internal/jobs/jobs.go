package jobs

import (
	t "github.com/faelmori/gastype/types"
	"github.com/google/uuid"
)

type Job struct {
	ID     string
	Action t.IAction
}

// NewJob creates a new job
func NewJob(action t.IAction) t.IJob {
	return &Job{
		ID:     uuid.NewString(),
		Action: action,
	}
}

// GetID returns the job ID
func (j *Job) GetID() string { return j.ID }

// GetAction returns the action associated with the job
func (j *Job) GetAction() t.IAction             { return j.Action }
func (j *Job) GetType() string                  { return j.Action.GetType() }
func (j *Job) GetResults() map[string]t.IResult { return j.Action.GetResults() }
func (j *Job) GetStatus() string                { return j.Action.GetStatus() }
func (j *Job) GetErrors() []error               { return j.Action.GetErrors() }
func (j *Job) IsRunning() bool                  { return j.Action.IsRunning() }
func (j *Job) CanExecute() bool                 { return j.Action.CanExecute() }
func (j *Job) Execute() error {
	if err := j.Action.Execute(); err != nil {
		return err
	}
	return nil
}
