package types

import (
	"math/big"
	"time"
)

type TaskResponse struct {
	TaskID        string
	OperatorAddr  string
	Signature     []byte
	Value         *big.Int
	BlockNumber   *big.Int
	ChainID       uint64
	TargetAddress string
	Key           *big.Int
	Epoch         uint32
	CreatedAt     time.Time
	Timestamp     *big.Int
} 