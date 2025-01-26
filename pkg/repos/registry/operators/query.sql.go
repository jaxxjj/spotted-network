// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package operators

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createOperator = `-- name: CreateOperator :one
INSERT INTO operators (
    address,
    signing_key,
    registered_at_block_number,
    registered_at_timestamp,
    active_epoch,
    status,
    weight
) VALUES ($1, $2, $3, $4, $5, 'inactive', $6)
RETURNING address, signing_key, registered_at_block_number, registered_at_timestamp, active_epoch, exit_epoch, status, weight, created_at, updated_at
`

type CreateOperatorParams struct {
	Address                 string         `json:"address"`
	SigningKey              string         `json:"signing_key"`
	RegisteredAtBlockNumber pgtype.Numeric `json:"registered_at_block_number"`
	RegisteredAtTimestamp   pgtype.Numeric `json:"registered_at_timestamp"`
	ActiveEpoch             pgtype.Numeric `json:"active_epoch"`
	Weight                  pgtype.Numeric `json:"weight"`
}

// Create a new operator record with inactive status
func (q *Queries) CreateOperator(ctx context.Context, arg CreateOperatorParams) (Operators, error) {
	row := q.db.QueryRow(ctx, createOperator,
		arg.Address,
		arg.SigningKey,
		arg.RegisteredAtBlockNumber,
		arg.RegisteredAtTimestamp,
		arg.ActiveEpoch,
		arg.Weight,
	)
	var i Operators
	err := row.Scan(
		&i.Address,
		&i.SigningKey,
		&i.RegisteredAtBlockNumber,
		&i.RegisteredAtTimestamp,
		&i.ActiveEpoch,
		&i.ExitEpoch,
		&i.Status,
		&i.Weight,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getOperatorByAddress = `-- name: GetOperatorByAddress :one
SELECT address, signing_key, registered_at_block_number, registered_at_timestamp, active_epoch, exit_epoch, status, weight, created_at, updated_at FROM operators
WHERE address = $1
`

// Get operator information by address
func (q *Queries) GetOperatorByAddress(ctx context.Context, address string) (Operators, error) {
	row := q.db.QueryRow(ctx, getOperatorByAddress, address)
	var i Operators
	err := row.Scan(
		&i.Address,
		&i.SigningKey,
		&i.RegisteredAtBlockNumber,
		&i.RegisteredAtTimestamp,
		&i.ActiveEpoch,
		&i.ExitEpoch,
		&i.Status,
		&i.Weight,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listAllOperators = `-- name: ListAllOperators :many
SELECT address, signing_key, registered_at_block_number, registered_at_timestamp, active_epoch, exit_epoch, status, weight, created_at, updated_at FROM operators
ORDER BY created_at DESC
`

// Get all operators regardless of status
func (q *Queries) ListAllOperators(ctx context.Context) ([]Operators, error) {
	rows, err := q.db.Query(ctx, listAllOperators)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Operators{}
	for rows.Next() {
		var i Operators
		if err := rows.Scan(
			&i.Address,
			&i.SigningKey,
			&i.RegisteredAtBlockNumber,
			&i.RegisteredAtTimestamp,
			&i.ActiveEpoch,
			&i.ExitEpoch,
			&i.Status,
			&i.Weight,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listOperatorsByStatus = `-- name: ListOperatorsByStatus :many
SELECT address, signing_key, registered_at_block_number, registered_at_timestamp, active_epoch, exit_epoch, status, weight, created_at, updated_at FROM operators
WHERE status = $1
ORDER BY created_at DESC
`

// Get all operators with a specific status
func (q *Queries) ListOperatorsByStatus(ctx context.Context, status string) ([]Operators, error) {
	rows, err := q.db.Query(ctx, listOperatorsByStatus, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Operators{}
	for rows.Next() {
		var i Operators
		if err := rows.Scan(
			&i.Address,
			&i.SigningKey,
			&i.RegisteredAtBlockNumber,
			&i.RegisteredAtTimestamp,
			&i.ActiveEpoch,
			&i.ExitEpoch,
			&i.Status,
			&i.Weight,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateOperatorExitEpoch = `-- name: UpdateOperatorExitEpoch :one
UPDATE operators
SET exit_epoch = $2,
    updated_at = NOW()
WHERE address = $1
RETURNING address, signing_key, registered_at_block_number, registered_at_timestamp, active_epoch, exit_epoch, status, weight, created_at, updated_at
`

type UpdateOperatorExitEpochParams struct {
	Address   string         `json:"address"`
	ExitEpoch pgtype.Numeric `json:"exit_epoch"`
}

// Update operator exit epoch
func (q *Queries) UpdateOperatorExitEpoch(ctx context.Context, arg UpdateOperatorExitEpochParams) (Operators, error) {
	row := q.db.QueryRow(ctx, updateOperatorExitEpoch, arg.Address, arg.ExitEpoch)
	var i Operators
	err := row.Scan(
		&i.Address,
		&i.SigningKey,
		&i.RegisteredAtBlockNumber,
		&i.RegisteredAtTimestamp,
		&i.ActiveEpoch,
		&i.ExitEpoch,
		&i.Status,
		&i.Weight,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateOperatorState = `-- name: UpdateOperatorState :one
UPDATE operators
SET status = $2,
    weight = $3,
    updated_at = NOW()
WHERE address = $1
RETURNING address, signing_key, registered_at_block_number, registered_at_timestamp, active_epoch, exit_epoch, status, weight, created_at, updated_at
`

type UpdateOperatorStateParams struct {
	Address string         `json:"address"`
	Status  string         `json:"status"`
	Weight  pgtype.Numeric `json:"weight"`
}

// Update operator status and weight
func (q *Queries) UpdateOperatorState(ctx context.Context, arg UpdateOperatorStateParams) (Operators, error) {
	row := q.db.QueryRow(ctx, updateOperatorState, arg.Address, arg.Status, arg.Weight)
	var i Operators
	err := row.Scan(
		&i.Address,
		&i.SigningKey,
		&i.RegisteredAtBlockNumber,
		&i.RegisteredAtTimestamp,
		&i.ActiveEpoch,
		&i.ExitEpoch,
		&i.Status,
		&i.Weight,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateOperatorStatus = `-- name: UpdateOperatorStatus :one
UPDATE operators
SET status = $2,
    updated_at = NOW()
WHERE address = $1
RETURNING address, signing_key, registered_at_block_number, registered_at_timestamp, active_epoch, exit_epoch, status, weight, created_at, updated_at
`

type UpdateOperatorStatusParams struct {
	Address string `json:"address"`
	Status  string `json:"status"`
}

// Update operator status
func (q *Queries) UpdateOperatorStatus(ctx context.Context, arg UpdateOperatorStatusParams) (Operators, error) {
	row := q.db.QueryRow(ctx, updateOperatorStatus, arg.Address, arg.Status)
	var i Operators
	err := row.Scan(
		&i.Address,
		&i.SigningKey,
		&i.RegisteredAtBlockNumber,
		&i.RegisteredAtTimestamp,
		&i.ActiveEpoch,
		&i.ExitEpoch,
		&i.Status,
		&i.Weight,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const upsertOperator = `-- name: UpsertOperator :one
INSERT INTO operators (
    address,
    signing_key,
    registered_at_block_number,
    registered_at_timestamp,
    active_epoch,
    status,
    weight,
    exit_epoch
) VALUES ($1, $2, $3, $4, $5, 'inactive', $6, $7)
ON CONFLICT (address) DO UPDATE
SET signing_key = EXCLUDED.signing_key,
    registered_at_block_number = EXCLUDED.registered_at_block_number,
    registered_at_timestamp = EXCLUDED.registered_at_timestamp,
    active_epoch = EXCLUDED.active_epoch,
    status = EXCLUDED.status,
    weight = EXCLUDED.weight,
    exit_epoch = EXCLUDED.exit_epoch,
    updated_at = NOW()
RETURNING address, signing_key, registered_at_block_number, registered_at_timestamp, active_epoch, exit_epoch, status, weight, created_at, updated_at
`

type UpsertOperatorParams struct {
	Address                 string         `json:"address"`
	SigningKey              string         `json:"signing_key"`
	RegisteredAtBlockNumber pgtype.Numeric `json:"registered_at_block_number"`
	RegisteredAtTimestamp   pgtype.Numeric `json:"registered_at_timestamp"`
	ActiveEpoch             pgtype.Numeric `json:"active_epoch"`
	Weight                  pgtype.Numeric `json:"weight"`
	ExitEpoch               pgtype.Numeric `json:"exit_epoch"`
}

// Insert or update operator record
func (q *Queries) UpsertOperator(ctx context.Context, arg UpsertOperatorParams) (Operators, error) {
	row := q.db.QueryRow(ctx, upsertOperator,
		arg.Address,
		arg.SigningKey,
		arg.RegisteredAtBlockNumber,
		arg.RegisteredAtTimestamp,
		arg.ActiveEpoch,
		arg.Weight,
		arg.ExitEpoch,
	)
	var i Operators
	err := row.Scan(
		&i.Address,
		&i.SigningKey,
		&i.RegisteredAtBlockNumber,
		&i.RegisteredAtTimestamp,
		&i.ActiveEpoch,
		&i.ExitEpoch,
		&i.Status,
		&i.Weight,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const verifyOperatorStatus = `-- name: VerifyOperatorStatus :one
SELECT status, signing_key 
FROM operators
WHERE address = $1 AND status = 'active'
`

type VerifyOperatorStatusRow struct {
	Status     string `json:"status"`
	SigningKey string `json:"signing_key"`
}

// Verify operator status and signing key for join request
func (q *Queries) VerifyOperatorStatus(ctx context.Context, address string) (VerifyOperatorStatusRow, error) {
	row := q.db.QueryRow(ctx, verifyOperatorStatus, address)
	var i VerifyOperatorStatusRow
	err := row.Scan(&i.Status, &i.SigningKey)
	return i, err
}
