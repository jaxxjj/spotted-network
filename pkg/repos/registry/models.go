// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package registry

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Operators struct {
	Address                 string           `json:"address"`
	SigningKey              string           `json:"signing_key"`
	RegisteredAtBlockNumber pgtype.Numeric   `json:"registered_at_block_number"`
	RegisteredAtTimestamp   pgtype.Numeric   `json:"registered_at_timestamp"`
	ActiveEpoch             int32            `json:"active_epoch"`
	ExitEpoch               pgtype.Int4      `json:"exit_epoch"`
	Status                  string           `json:"status"`
	Weight                  pgtype.Numeric   `json:"weight"`
	Missing                 pgtype.Int4      `json:"missing"`
	SuccessfulResponseCount pgtype.Int4      `json:"successful_response_count"`
	CreatedAt               pgtype.Timestamp `json:"created_at"`
	UpdatedAt               pgtype.Timestamp `json:"updated_at"`
}
