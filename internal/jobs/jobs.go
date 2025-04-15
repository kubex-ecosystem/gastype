package jobs

import (
	"fmt"
	t "github.com/faelmori/gastype/types"
	l "github.com/faelmori/logz"
	"github.com/google/uuid"
	"time"
)

type Job struct {
	logger       l.Logger
	ID           string
	CreateTime   string
	CancelTime   time.Time
	FinishTime   time.Time
	Results      map[string]t.IResult
	Errors       []error
	Running      bool
	CancelChanel chan struct{}
	DoneChanel   chan struct{}
	Action       t.IAction
}

// NewJob creates a new job
func NewJob(action t.IAction, cancelChanel chan struct{}, doneChanel chan struct{}, logger l.Logger) *Job {
	if logger == nil {
		logger = l.GetLogger("GasType")
	}
	if cancelChanel == nil {
		cancelChanel = make(chan struct{}, 2)
	}
	if doneChanel == nil {
		doneChanel = make(chan struct{}, 2)
	}
	return &Job{
		logger:       logger,
		ID:           uuid.NewString(),
		Results:      make(map[string]t.IResult),
		Errors:       make([]error, 0),
		Running:      false,
		CancelChanel: cancelChanel,
		DoneChanel:   doneChanel,
		FinishTime:   time.Time{},
		CancelTime:   time.Time{},
		CreateTime:   time.Now().Format(time.RFC3339),
		Action:       action,
	}
}

// GetID returns the job ID
func (j *Job) GetID() string {
	return j.ID
}

// GetAction returns the action associated with the job
func (j *Job) GetAction() t.IAction {
	return j.Action
}
func (j *Job) GetType() string {
	return j.Action.GetType()
}
func (j *Job) GetResults() map[string]t.IResult {
	return j.Action.GetResults()
}
func (j *Job) GetStatus() string {
	return j.Action.GetStatus()
}
func (j *Job) GetErrors() []error {
	return j.Action.GetErrors()
}
func (j *Job) IsRunning() bool {
	return j.Action.IsRunning()
}
func (j *Job) CanExecute() bool {
	return j.Action.CanExecute()
}
func (j *Job) Execute() error {
	if j.Running {
		j.logger.ErrorCtx("Job is already running", map[string]interface{}{"job_id": j.ID})
		return nil
	}
	if j.Action.CanExecute() {
		if err := j.Action.Execute(); err != nil {
			return err
		}
	} else {
		j.logger.ErrorCtx("Job cannot be executed", map[string]interface{}{"job_id": j.ID})
		return nil
	}
	return nil
}

func (j *Job) GetErrorChannel() chan error {
	return j.Action.GetErrorChannel()
}
func (j *Job) GetDoneChannel() chan struct{} {
	return j.Action.GetDoneChannel()
}
func (j *Job) GetCancelChannel() chan struct{} {
	return j.Action.GetCancelChannel()
}
func (j *Job) GetResultsChannel() chan t.IResult {
	return j.Action.GetResultsChannel()
}

func (j *Job) GetCreateTime() time.Time {
	createTime, err := time.Parse(time.RFC3339, j.CreateTime)
	if err != nil {
		j.logger.ErrorCtx("Error parsing create time", map[string]interface{}{"job_id": j.ID, "error": err})
		return time.Time{}
	}
	return createTime
}
func (j *Job) GetFinishTime() time.Time {
	return j.FinishTime
}
func (j *Job) GetCancelTime() time.Time {
	return j.CancelTime
}
func (j *Job) SetFinishTime(t time.Time) {
	j.FinishTime = t
}
func (j *Job) SetCancelTime(t time.Time) {
	j.CancelTime = t
}
func (j *Job) Cancel() error {
	if j.Running {
		ch := j.GetCancelChannel()
		if ch != nil {
			defer close(ch)
			ch <- struct{}{}
		}
		j.Running = false
		j.CancelTime = time.Now()
		j.logger.InfoCtx("Job cancelled", map[string]interface{}{"job_id": j.ID})
		return nil
	} else {
		j.logger.ErrorCtx("Job is not running", map[string]interface{}{"job_id": j.ID})
		return fmt.Errorf("job is not running")
	}
}
