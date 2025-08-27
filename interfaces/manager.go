package interfaces

type ITypeManager interface {
	GetNotifierChan() chan string
	GetEmail() string
	GetEmailToken() string
	GetNotify() bool
	GetConfig() IConfig
	GetActions() []IAction
	IsRunning() bool
	SetNotifierChan(notifierChan chan string)
	SetEmail(email string)
	SetEmailToken(emailToken string)
	SetNotify(notify bool)
	SetConfig(cfg IConfig)
	AddAction(action IAction)
	StartChecking(workerCount int) error
	StopChecking()
	LoadConfig() error
	SaveConfig() error
	CanNotify() bool
	PrepareActions() error
}
