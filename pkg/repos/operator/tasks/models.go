// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v2.3.0-wicked-fork

package tasks

import (
	"time"

	types "github.com/galxe/spotted-network/pkg/common/types"
	"github.com/jackc/pgx/v5/pgtype"
)

type Tasks struct {
	TaskID                string           `json:"task_id"`
	ChainID               uint32           `json:"chain_id"`
	TargetAddress         string           `json:"target_address"`
	Key                   pgtype.Numeric   `json:"key"`
	BlockNumber           uint64           `json:"block_number"`
	Value                 pgtype.Numeric   `json:"value"`
	Epoch                 uint32           `json:"epoch"`
	Status                types.TaskStatus `json:"status"`
	RequiredConfirmations uint16           `json:"required_confirmations"`
	RetryCount            uint16           `json:"retry_count"`
	CreatedAt             time.Time        `json:"created_at"`
	UpdatedAt             time.Time        `json:"updated_at"`
}
