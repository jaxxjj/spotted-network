// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package consensus_responses

import (
	"context"
)

type Querier interface {
	// -- invalidate: GetConsensusResponseByTaskId
	// -- invalidate: GetConsensusResponseByRequest
	CreateConsensusResponse(ctx context.Context, arg CreateConsensusResponseParams) (ConsensusResponse, error)
	// -- cache: 7d
	GetConsensusResponseByRequest(ctx context.Context, arg GetConsensusResponseByRequestParams) (ConsensusResponse, error)
	// -- cache: 7d
	GetConsensusResponseByTaskId(ctx context.Context, taskID string) (ConsensusResponse, error)
}

var _ Querier = (*Queries)(nil)
