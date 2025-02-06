package constants


const (
	GenesisBlock = 0
	EpochPeriod = 12
)

type TaskStatus string

const (
    TaskStatusPending    TaskStatus = "pending"
    TaskStatusCompleted  TaskStatus = "completed"
    TaskStatusConfirming TaskStatus = "confirming"
)



