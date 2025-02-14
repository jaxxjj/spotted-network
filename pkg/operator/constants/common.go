package constants

const (
	GenesisBlock = 7698497
	EpochPeriod  = 12
	GracePeriod  = 3
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusConfirming TaskStatus = "confirming"
)
