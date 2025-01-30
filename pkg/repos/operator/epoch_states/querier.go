// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v2.3.0-wicked-fork

package epoch_states

import (
	"context"
)

// TODO: change things here.
type Querier interface {
	// -- cache: 168h
	// -- timeout: 500ms
	GetEpochState(ctx context.Context, epochNumber uint32) (*EpochState, error)
	// -- cache: 1h
	// -- timeout: 500ms
	GetLatestEpochState(ctx context.Context) (*EpochState, error)
	// -- cache: 168h
	// -- timeout: 1s
	ListEpochStates(ctx context.Context, limit int32) ([]EpochState, error)
	// -- invalidate: GetEpochState
	// -- invalidate: GetLatestEpochState
	// -- invalidate: ListEpochStates
	// -- timeout: 500ms
	UpsertEpochState(ctx context.Context, arg UpsertEpochStateParams, listEpochStates *int32) (*EpochState, error)
}

var _ Querier = (*Queries)(nil)
