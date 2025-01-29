package types

// Task represents a state query task

type TaskStatus string

const (
    TaskStatusPending    TaskStatus = "pending"
    TaskStatusCompleted  TaskStatus = "completed"
    TaskStatusFailed     TaskStatus = "failed"
    TaskStatusConfirming TaskStatus = "confirming"
)

// OperatorStatus represents the status of an operator
type OperatorStatus string

const (
	OperatorStatusActive OperatorStatus = "active"
	OperatorStatusInactive OperatorStatus = "inactive"
	OperatorStatusSuspended OperatorStatus = "suspended"
)

