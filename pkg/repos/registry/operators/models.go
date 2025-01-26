// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package operators

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Operators struct {
	Address                 string             `json:"address"`
	SigningKey              string             `json:"signing_key"`
	RegisteredAtBlockNumber pgtype.Numeric     `json:"registered_at_block_number"`
	RegisteredAtTimestamp   pgtype.Numeric     `json:"registered_at_timestamp"`
	ActiveEpoch             pgtype.Numeric     `json:"active_epoch"`
	ExitEpoch               pgtype.Numeric     `json:"exit_epoch"`
	Status                  string             `json:"status"`
	Weight                  pgtype.Numeric     `json:"weight"`
	CreatedAt               pgtype.Timestamptz `json:"created_at"`
	UpdatedAt               pgtype.Timestamptz `json:"updated_at"`
}
