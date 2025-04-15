package jobs

import (
	"fmt"
	t "github.com/faelmori/gastype/types"
	l "github.com/faelmori/logz"
	"github.com/google/uuid"
	"time"
)

// Job represents a job that executes an action and tracks its state and results.
type Job struct {
	logger       l.Logger             // Logger instance for logging job-related information.
	ID           string               // Unique identifier for the job.
	CreateTime   string               // Creation time of the job in RFC3339 format.
	CancelTime   time.Time            // Time when the job was canceled.
	FinishTime   time.Time            // Time when the job was finished.
	Results      map[string]t.IResult // Map of results associated with the job.
	Errors       []error              // List of errors encountered during the job execution.
	Running      bool                 // Indicates whether the job is currently running.
	CancelChanel chan struct{}        // Channel used to signal job cancellation.
	DoneChanel   chan struct{}        // Channel used to signal job completion.
	Action       t.IAction            // Action associated with the job.
}

// NewJob creates a new job.
// Parameters:
//   - action: The action to be executed by the job.
//   - cancelChanel: Channel to signal job cancellation.
//   - doneChanel: Channel to signal job completion.
//   - logger: Logger instance for the job.
//
// Returns:
//   - *Job: A new instance of the Job struct.
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

// GetID returns the job ID.
// Returns:
//   - string: The unique identifier of the job.
func (j *Job) GetID() string {
	return j.ID
}

// GetAction returns the action associated with the job.
// Returns:
//   - t.IAction: The action instance.
func (j *Job) GetAction() t.IAction {
	return j.Action
}

// GetType returns the type of the action associated with the job.
// Returns:
//   - string: The type of the action.
func (j *Job) GetType() string {
	return j.Action.GetType()
}

// GetResults returns the results of the action associated with the job.
// Returns:
//   - map[string]t.IResult: A map of results.
func (j *Job) GetResults() map[string]t.IResult {
	return j.Action.GetResults()
}

// GetStatus returns the current status of the action associated with the job.
// Returns:
//   - string: The status of the action.
func (j *Job) GetStatus() string {
	return j.Action.GetStatus()
}

// GetErrors returns the errors encountered during the action execution.
// Returns:
//   - []error: A list of errors.
func (j *Job) GetErrors() []error {
	return j.Action.GetErrors()
}

// IsRunning checks if the action associated with the job is currently running.
// Returns:
//   - bool: True if the action is running, false otherwise.
func (j *Job) IsRunning() bool {
	return j.Action.IsRunning()
}

// CanExecute checks if the action associated with the job can be executed.
// Returns:
//   - bool: True if the action can be executed, false otherwise.
func (j *Job) CanExecute() bool {
	return j.Action.CanExecute()
}

// Execute starts the execution of the action associated with the job.
// Returns:
//   - error: An error if the job is already running or cannot be executed, nil otherwise.
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

// GetErrorChannel returns the error channel of the action associated with the job.
// Returns:
//   - chan error: The error channel.
func (j *Job) GetErrorChannel() chan error {
	return j.Action.GetErrorChannel()
}

// GetDoneChannel returns the done channel of the action associated with the job.
// Returns:
//   - chan struct{}: The done channel.
func (j *Job) GetDoneChannel() chan struct{} {
	return j.Action.GetDoneChannel()
}

// GetCancelChannel returns the cancel channel of the action associated with the job.
// Returns:
//   - chan struct{}: The cancel channel.
func (j *Job) GetCancelChannel() chan struct{} {
	return j.Action.GetCancelChannel()
}

// GetResultsChannel returns the results channel of the action associated with the job.
// Returns:
//   - chan t.IResult: The results channel.
func (j *Job) GetResultsChannel() chan t.IResult {
	return j.Action.GetResultsChannel()
}

// GetCreateTime returns the creation time of the job.
// Returns:
//   - time.Time: The creation time.
func (j *Job) GetCreateTime() time.Time {
	createTime, err := time.Parse(time.RFC3339, j.CreateTime)
	if err != nil {
		j.logger.ErrorCtx("Error parsing create time", map[string]interface{}{"job_id": j.ID, "error": err})
		return time.Time{}
	}
	return createTime
}

// GetFinishTime returns the finish time of the job.
// Returns:
//   - time.Time: The finish time.
func (j *Job) GetFinishTime() time.Time {
	return j.FinishTime
}

// GetCancelTime returns the cancel time of the job.
// Returns:
//   - time.Time: The cancel time.
func (j *Job) GetCancelTime() time.Time {
	return j.CancelTime
}

// SetFinishTime sets the finish time of the job.
// Parameters:
//   - t: The finish time to set.
func (j *Job) SetFinishTime(t time.Time) {
	j.FinishTime = t
}

// SetCancelTime sets the cancel time of the job.
// Parameters:
//   - t: The cancel time to set.
func (j *Job) SetCancelTime(t time.Time) {
	j.CancelTime = t
}

// Cancel cancels the job and updates its state.
// Returns:
//   - error: An error if the job is not running, nil otherwise.
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
