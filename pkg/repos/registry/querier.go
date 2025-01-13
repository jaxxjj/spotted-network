// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package registry

import (
	"context"
)

type Querier interface {
	// Create a new operator record with waitingJoin status
	CreateOperator(ctx context.Context, arg CreateOperatorParams) (Operators, error)
	// Get operator information by address
	GetOperatorByAddress(ctx context.Context, address string) (Operators, error)
	// Get all operators with a specific status
	ListOperatorsByStatus(ctx context.Context, status string) ([]Operators, error)
	// Update operator status
	UpdateOperatorStatus(ctx context.Context, arg UpdateOperatorStatusParams) (Operators, error)
	// Verify operator status and signing key for join request
	VerifyOperatorStatus(ctx context.Context, address string) (VerifyOperatorStatusRow, error)
}

var _ Querier = (*Queries)(nil)
