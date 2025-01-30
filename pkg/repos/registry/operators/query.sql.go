// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v2.3.0-wicked-fork
// source: query.sql

package operators

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	types "github.com/galxe/spotted-network/pkg/common/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
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
	RegisteredAtBlockNumber uint64         `json:"registered_at_block_number"`
	RegisteredAtTimestamp   uint64         `json:"registered_at_timestamp"`
	ActiveEpoch             uint32         `json:"active_epoch"`
	Weight                  pgtype.Numeric `json:"weight"`
}

// Create a new operator record with inactive status
// -- invalidate: GetOperatorByAddress
// -- timeout: 500ms
func (q *Queries) CreateOperator(ctx context.Context, arg CreateOperatorParams, getOperatorByAddress *string) (*Operators, error) {
	return _CreateOperator(ctx, q, arg, getOperatorByAddress)
}

func _CreateOperator(ctx context.Context, q CacheWGConn, arg CreateOperatorParams, getOperatorByAddress *string) (*Operators, error) {
	qctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	row := q.GetConn().WQueryRow(qctx, "operators.CreateOperator", createOperator,
		arg.Address,
		arg.SigningKey,
		arg.RegisteredAtBlockNumber,
		arg.RegisteredAtTimestamp,
		arg.ActiveEpoch,
		arg.Weight)
	var i *Operators = new(Operators)
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
	if err == pgx.ErrNoRows {
		return (*Operators)(nil), nil
	} else if err != nil {
		return nil, err
	}

	// invalidate
	_ = q.GetConn().PostExec(func() error {
		anyErr := make(chan error, 1)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			if getOperatorByAddress != nil {
				key := "operators:GetOperatorByAddress:" + hashIfLong(fmt.Sprintf("%+v", (*getOperatorByAddress)))
				err = q.GetCache().Invalidate(ctx, key)
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Msgf(
						"Failed to invalidate: %s", key)
					anyErr <- err
				}
			}
		}()
		wg.Wait()
		close(anyErr)
		return <-anyErr
	})
	return i, err
}

const getOperatorByAddress = `-- name: GetOperatorByAddress :one
SELECT address, signing_key, registered_at_block_number, registered_at_timestamp, active_epoch, exit_epoch, status, weight, created_at, updated_at FROM operators
WHERE address = $1
`

// Get operator information by address
// -- cache: 168h
// -- timeout: 500ms
func (q *Queries) GetOperatorByAddress(ctx context.Context, address string) (*Operators, error) {
	return _GetOperatorByAddress(ctx, q.AsReadOnly(), address)
}

func (q *ReadOnlyQueries) GetOperatorByAddress(ctx context.Context, address string) (*Operators, error) {
	return _GetOperatorByAddress(ctx, q, address)
}

func _GetOperatorByAddress(ctx context.Context, q CacheQuerierConn, address string) (*Operators, error) {
	qctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	q.GetConn().CountIntent("operators.GetOperatorByAddress")
	dbRead := func() (any, time.Duration, error) {
		cacheDuration := time.Duration(time.Millisecond * 604800000)
		row := q.GetConn().WQueryRow(qctx, "operators.GetOperatorByAddress", getOperatorByAddress, address)
		var i *Operators = new(Operators)
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
		if err == pgx.ErrNoRows {
			return (*Operators)(nil), cacheDuration, nil
		}
		return i, cacheDuration, err
	}
	if q.GetCache() == nil {
		i, _, err := dbRead()
		return i.(*Operators), err
	}

	var i *Operators
	err := q.GetCache().GetWithTtl(qctx, "operators:GetOperatorByAddress:"+hashIfLong(fmt.Sprintf("%+v", address)), &i, dbRead, false, false)
	if err != nil {
		return nil, err
	}

	return i, err
}

const listAllOperators = `-- name: ListAllOperators :many
SELECT address, signing_key, registered_at_block_number, registered_at_timestamp, active_epoch, exit_epoch, status, weight, created_at, updated_at FROM operators
ORDER BY created_at DESC
`

// -- timeout: 1s
func (q *Queries) ListAllOperators(ctx context.Context) ([]Operators, error) {
	return _ListAllOperators(ctx, q.AsReadOnly())
}

func (q *ReadOnlyQueries) ListAllOperators(ctx context.Context) ([]Operators, error) {
	return _ListAllOperators(ctx, q)
}

func _ListAllOperators(ctx context.Context, q CacheQuerierConn) ([]Operators, error) {
	qctx, cancel := context.WithTimeout(ctx, time.Millisecond*1000)
	defer cancel()
	q.GetConn().CountIntent("operators.ListAllOperators")
	rows, err := q.GetConn().WQuery(qctx, "operators.ListAllOperators", listAllOperators)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Operators
	for rows.Next() {
		var i *Operators = new(Operators)
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
		items = append(items, *i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, err
}

const listOperatorsByStatus = `-- name: ListOperatorsByStatus :many
SELECT address, signing_key, registered_at_block_number, registered_at_timestamp, active_epoch, exit_epoch, status, weight, created_at, updated_at FROM operators
WHERE status = $1
ORDER BY created_at DESC
`

// -- timeout: 1s
func (q *Queries) ListOperatorsByStatus(ctx context.Context, status types.OperatorStatus) ([]Operators, error) {
	return _ListOperatorsByStatus(ctx, q.AsReadOnly(), status)
}

func (q *ReadOnlyQueries) ListOperatorsByStatus(ctx context.Context, status types.OperatorStatus) ([]Operators, error) {
	return _ListOperatorsByStatus(ctx, q, status)
}

func _ListOperatorsByStatus(ctx context.Context, q CacheQuerierConn, status types.OperatorStatus) ([]Operators, error) {
	qctx, cancel := context.WithTimeout(ctx, time.Millisecond*1000)
	defer cancel()
	q.GetConn().CountIntent("operators.ListOperatorsByStatus")
	rows, err := q.GetConn().WQuery(qctx, "operators.ListOperatorsByStatus", listOperatorsByStatus, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Operators
	for rows.Next() {
		var i *Operators = new(Operators)
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
		items = append(items, *i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, err
}

const updateOperatorExitEpoch = `-- name: UpdateOperatorExitEpoch :one
UPDATE operators
SET exit_epoch = $2,
    updated_at = NOW()
WHERE address = $1
RETURNING address, signing_key, registered_at_block_number, registered_at_timestamp, active_epoch, exit_epoch, status, weight, created_at, updated_at
`

type UpdateOperatorExitEpochParams struct {
	Address   string `json:"address"`
	ExitEpoch uint32 `json:"exit_epoch"`
}

// -- invalidate: GetOperatorByAddress
// -- timeout: 500ms
func (q *Queries) UpdateOperatorExitEpoch(ctx context.Context, arg UpdateOperatorExitEpochParams, getOperatorByAddress *string) (*Operators, error) {
	return _UpdateOperatorExitEpoch(ctx, q, arg, getOperatorByAddress)
}

func _UpdateOperatorExitEpoch(ctx context.Context, q CacheWGConn, arg UpdateOperatorExitEpochParams, getOperatorByAddress *string) (*Operators, error) {
	qctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	row := q.GetConn().WQueryRow(qctx, "operators.UpdateOperatorExitEpoch", updateOperatorExitEpoch, arg.Address, arg.ExitEpoch)
	var i *Operators = new(Operators)
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
	if err == pgx.ErrNoRows {
		return (*Operators)(nil), nil
	} else if err != nil {
		return nil, err
	}

	// invalidate
	_ = q.GetConn().PostExec(func() error {
		anyErr := make(chan error, 1)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			if getOperatorByAddress != nil {
				key := "operators:GetOperatorByAddress:" + hashIfLong(fmt.Sprintf("%+v", (*getOperatorByAddress)))
				err = q.GetCache().Invalidate(ctx, key)
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Msgf(
						"Failed to invalidate: %s", key)
					anyErr <- err
				}
			}
		}()
		wg.Wait()
		close(anyErr)
		return <-anyErr
	})
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
	Address string               `json:"address"`
	Status  types.OperatorStatus `json:"status"`
	Weight  pgtype.Numeric       `json:"weight"`
}

// -- invalidate: GetOperatorByAddress
// -- timeout: 500ms
func (q *Queries) UpdateOperatorState(ctx context.Context, arg UpdateOperatorStateParams, getOperatorByAddress *string) (*Operators, error) {
	return _UpdateOperatorState(ctx, q, arg, getOperatorByAddress)
}

func _UpdateOperatorState(ctx context.Context, q CacheWGConn, arg UpdateOperatorStateParams, getOperatorByAddress *string) (*Operators, error) {
	qctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	row := q.GetConn().WQueryRow(qctx, "operators.UpdateOperatorState", updateOperatorState, arg.Address, arg.Status, arg.Weight)
	var i *Operators = new(Operators)
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
	if err == pgx.ErrNoRows {
		return (*Operators)(nil), nil
	} else if err != nil {
		return nil, err
	}

	// invalidate
	_ = q.GetConn().PostExec(func() error {
		anyErr := make(chan error, 1)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			if getOperatorByAddress != nil {
				key := "operators:GetOperatorByAddress:" + hashIfLong(fmt.Sprintf("%+v", (*getOperatorByAddress)))
				err = q.GetCache().Invalidate(ctx, key)
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Msgf(
						"Failed to invalidate: %s", key)
					anyErr <- err
				}
			}
		}()
		wg.Wait()
		close(anyErr)
		return <-anyErr
	})
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
	Address string               `json:"address"`
	Status  types.OperatorStatus `json:"status"`
}

// -- invalidate: GetOperatorByAddress
// -- timeout: 500ms
func (q *Queries) UpdateOperatorStatus(ctx context.Context, arg UpdateOperatorStatusParams, getOperatorByAddress *string) (*Operators, error) {
	return _UpdateOperatorStatus(ctx, q, arg, getOperatorByAddress)
}

func _UpdateOperatorStatus(ctx context.Context, q CacheWGConn, arg UpdateOperatorStatusParams, getOperatorByAddress *string) (*Operators, error) {
	qctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	row := q.GetConn().WQueryRow(qctx, "operators.UpdateOperatorStatus", updateOperatorStatus, arg.Address, arg.Status)
	var i *Operators = new(Operators)
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
	if err == pgx.ErrNoRows {
		return (*Operators)(nil), nil
	} else if err != nil {
		return nil, err
	}

	// invalidate
	_ = q.GetConn().PostExec(func() error {
		anyErr := make(chan error, 1)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			if getOperatorByAddress != nil {
				key := "operators:GetOperatorByAddress:" + hashIfLong(fmt.Sprintf("%+v", (*getOperatorByAddress)))
				err = q.GetCache().Invalidate(ctx, key)
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Msgf(
						"Failed to invalidate: %s", key)
					anyErr <- err
				}
			}
		}()
		wg.Wait()
		close(anyErr)
		return <-anyErr
	})
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
	RegisteredAtBlockNumber uint64         `json:"registered_at_block_number"`
	RegisteredAtTimestamp   uint64         `json:"registered_at_timestamp"`
	ActiveEpoch             uint32         `json:"active_epoch"`
	Weight                  pgtype.Numeric `json:"weight"`
	ExitEpoch               uint32         `json:"exit_epoch"`
}

// -- invalidate: GetOperatorByAddress
// -- timeout: 500ms
func (q *Queries) UpsertOperator(ctx context.Context, arg UpsertOperatorParams, getOperatorByAddress *string) (*Operators, error) {
	return _UpsertOperator(ctx, q, arg, getOperatorByAddress)
}

func _UpsertOperator(ctx context.Context, q CacheWGConn, arg UpsertOperatorParams, getOperatorByAddress *string) (*Operators, error) {
	qctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	row := q.GetConn().WQueryRow(qctx, "operators.UpsertOperator", upsertOperator,
		arg.Address,
		arg.SigningKey,
		arg.RegisteredAtBlockNumber,
		arg.RegisteredAtTimestamp,
		arg.ActiveEpoch,
		arg.Weight,
		arg.ExitEpoch)
	var i *Operators = new(Operators)
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
	if err == pgx.ErrNoRows {
		return (*Operators)(nil), nil
	} else if err != nil {
		return nil, err
	}

	// invalidate
	_ = q.GetConn().PostExec(func() error {
		anyErr := make(chan error, 1)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			if getOperatorByAddress != nil {
				key := "operators:GetOperatorByAddress:" + hashIfLong(fmt.Sprintf("%+v", (*getOperatorByAddress)))
				err = q.GetCache().Invalidate(ctx, key)
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Msgf(
						"Failed to invalidate: %s", key)
					anyErr <- err
				}
			}
		}()
		wg.Wait()
		close(anyErr)
		return <-anyErr
	})
	return i, err
}

const verifyOperatorStatus = `-- name: VerifyOperatorStatus :one
SELECT status, signing_key 
FROM operators
WHERE address = $1 AND status = 'active'
`

type VerifyOperatorStatusRow struct {
	Status     types.OperatorStatus `json:"status"`
	SigningKey string               `json:"signing_key"`
}

// -- timeout: 500ms
func (q *Queries) VerifyOperatorStatus(ctx context.Context, address string) (*VerifyOperatorStatusRow, error) {
	return _VerifyOperatorStatus(ctx, q.AsReadOnly(), address)
}

func (q *ReadOnlyQueries) VerifyOperatorStatus(ctx context.Context, address string) (*VerifyOperatorStatusRow, error) {
	return _VerifyOperatorStatus(ctx, q, address)
}

func _VerifyOperatorStatus(ctx context.Context, q CacheQuerierConn, address string) (*VerifyOperatorStatusRow, error) {
	qctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	q.GetConn().CountIntent("operators.VerifyOperatorStatus")
	row := q.GetConn().WQueryRow(qctx, "operators.VerifyOperatorStatus", verifyOperatorStatus, address)
	var i *VerifyOperatorStatusRow = new(VerifyOperatorStatusRow)
	err := row.Scan(&i.Status, &i.SigningKey)
	if err == pgx.ErrNoRows {
		return (*VerifyOperatorStatusRow)(nil), nil
	} else if err != nil {
		return nil, err
	}

	return i, err
}

//// auto generated functions

func (q *Queries) Dump(ctx context.Context, beforeDump ...BeforeDump) ([]byte, error) {
	sql := "SELECT address,signing_key,registered_at_block_number,registered_at_timestamp,active_epoch,exit_epoch,status,weight,created_at,updated_at FROM \"operators\" ORDER BY address,signing_key,created_at,updated_at ASC;"
	rows, err := q.db.WQuery(ctx, "operators.Dump", sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Operators
	for rows.Next() {
		var v Operators
		if err := rows.Scan(&v.Address, &v.SigningKey, &v.RegisteredAtBlockNumber, &v.RegisteredAtTimestamp, &v.ActiveEpoch, &v.ExitEpoch, &v.Status, &v.Weight, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		for _, applyBeforeDump := range beforeDump {
			applyBeforeDump(&v)
		}
		items = append(items, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	bytes, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (q *Queries) Load(ctx context.Context, data []byte) error {
	sql := "INSERT INTO \"operators\" (address,signing_key,registered_at_block_number,registered_at_timestamp,active_epoch,exit_epoch,status,weight,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10);"
	rows := make([]Operators, 0)
	err := json.Unmarshal(data, &rows)
	if err != nil {
		return err
	}
	for _, row := range rows {
		_, err := q.db.WExec(ctx, "operators.Load", sql, row.Address, row.SigningKey, row.RegisteredAtBlockNumber, row.RegisteredAtTimestamp, row.ActiveEpoch, row.ExitEpoch, row.Status, row.Weight, row.CreatedAt, row.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func hashIfLong(v string) string {
	if len(v) > 64 {
		hash := sha256.Sum256([]byte(v))
		return "h(" + hex.EncodeToString(hash[:]) + ")"
	}
	return v
}

func ptrStr[T any](v *T) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%+v", *v)
}

// eliminate unused error
var _ = log.Logger
var _ = fmt.Sprintf("")
var _ = time.Now()
var _ = json.RawMessage{}
var _ = sha256.Sum256(nil)
var _ = hex.EncodeToString(nil)
var _ = sync.WaitGroup{}
