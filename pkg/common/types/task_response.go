package types

import (
	"math/big"
	"time"
)

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