// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package epoch_states

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type EpochState struct {
	EpochNumber     uint32             `json:"epoch_number"`
	BlockNumber     uint64             `json:"block_number"`
	MinimumWeight   pgtype.Numeric     `json:"minimum_weight"`
	TotalWeight     pgtype.Numeric     `json:"total_weight"`
	ThresholdWeight pgtype.Numeric     `json:"threshold_weight"`
	UpdatedAt       pgtype.Timestamptz `json:"updated_at"`
}
