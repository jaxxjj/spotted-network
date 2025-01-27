// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package consensus_responses

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type ConsensusResponse struct {
	ID                   int64              `json:"id"`
	TaskID               string             `json:"task_id"`
	Epoch                uint32             `json:"epoch"`
	Value                pgtype.Numeric     `json:"value"`
	Key                  pgtype.Numeric     `json:"key"`
	TotalWeight          pgtype.Numeric     `json:"total_weight"`
	ChainID              uint32             `json:"chain_id"`
	BlockNumber          uint64             `json:"block_number"`
	TargetAddress        string             `json:"target_address"`
	AggregatedSignatures []byte             `json:"aggregated_signatures"`
	OperatorSignatures   []byte             `json:"operator_signatures"`
	ConsensusReachedAt   pgtype.Timestamptz `json:"consensus_reached_at"`
	CreatedAt            pgtype.Timestamptz `json:"created_at"`
	UpdatedAt            pgtype.Timestamptz `json:"updated_at"`
}
