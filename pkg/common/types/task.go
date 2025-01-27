package types

import (
	"math/big"
	"time"
)

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

type Task struct {
	ID            string
	ChainID       uint64
	TargetAddress string
	Key           *big.Int
	Value         *big.Int
	BlockNumber   *big.Int
	Timestamp     *big.Int
	Epoch         uint32
	CreatedAt     time.Time
}

// TaskResponse represents a response from an operator for a task
type TaskResponse struct {
	TaskID        string
	OperatorAddr  string
	SigningKey    string
	Signature     []byte
	Value         *big.Int
	BlockNumber   *big.Int
	ChainID       uint64
	TargetAddress string
	Key           *big.Int
	Epoch         uint32
	Timestamp     *big.Int
	CreatedAt     time.Time
}
