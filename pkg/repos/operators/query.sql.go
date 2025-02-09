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

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

const getOperatorByAddress = `-- name: GetOperatorByAddress :one
SELECT address, signing_key, p2p_key, registered_at_block_number, active_epoch, exit_epoch, is_active, weight, created_at, updated_at FROM operators
WHERE address = $1
`

// -- cache: 30s
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
		cacheDuration := time.Duration(time.Millisecond * 30000)
		row := q.GetConn().WQueryRow(qctx, "operators.GetOperatorByAddress", getOperatorByAddress, address)
		var i *Operators = new(Operators)
		err := row.Scan(
			&i.Address,
			&i.SigningKey,
			&i.P2pKey,
			&i.RegisteredAtBlockNumber,
			&i.ActiveEpoch,
			&i.ExitEpoch,
			&i.IsActive,
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

const getOperatorByP2PKey = `-- name: GetOperatorByP2PKey :one
SELECT address, signing_key, p2p_key, registered_at_block_number, active_epoch, exit_epoch, is_active, weight, created_at, updated_at FROM operators
WHERE LOWER(p2p_key) = LOWER($1)
`

// -- cache: 30s
// -- timeout: 500ms
func (q *Queries) GetOperatorByP2PKey(ctx context.Context, lower string) (*Operators, error) {
	return _GetOperatorByP2PKey(ctx, q.AsReadOnly(), lower)
}

func (q *ReadOnlyQueries) GetOperatorByP2PKey(ctx context.Context, lower string) (*Operators, error) {
	return _GetOperatorByP2PKey(ctx, q, lower)
}

func _GetOperatorByP2PKey(ctx context.Context, q CacheQuerierConn, lower string) (*Operators, error) {
	qctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	q.GetConn().CountIntent("operators.GetOperatorByP2PKey")
	dbRead := func() (any, time.Duration, error) {
		cacheDuration := time.Duration(time.Millisecond * 30000)
		row := q.GetConn().WQueryRow(qctx, "operators.GetOperatorByP2PKey", getOperatorByP2PKey, lower)
		var i *Operators = new(Operators)
		err := row.Scan(
			&i.Address,
			&i.SigningKey,
			&i.P2pKey,
			&i.RegisteredAtBlockNumber,
			&i.ActiveEpoch,
			&i.ExitEpoch,
			&i.IsActive,
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
	err := q.GetCache().GetWithTtl(qctx, "operators:GetOperatorByP2PKey:"+hashIfLong(fmt.Sprintf("%+v", lower)), &i, dbRead, false, false)
	if err != nil {
		return nil, err
	}

	return i, err
}

const getOperatorBySigningKey = `-- name: GetOperatorBySigningKey :one
SELECT address, signing_key, p2p_key, registered_at_block_number, active_epoch, exit_epoch, is_active, weight, created_at, updated_at FROM operators
WHERE signing_key = $1
`

// -- cache: 30s
// -- timeout: 500ms
func (q *Queries) GetOperatorBySigningKey(ctx context.Context, signingKey string) (*Operators, error) {
	return _GetOperatorBySigningKey(ctx, q.AsReadOnly(), signingKey)
}

func (q *ReadOnlyQueries) GetOperatorBySigningKey(ctx context.Context, signingKey string) (*Operators, error) {
	return _GetOperatorBySigningKey(ctx, q, signingKey)
}

func _GetOperatorBySigningKey(ctx context.Context, q CacheQuerierConn, signingKey string) (*Operators, error) {
	qctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	q.GetConn().CountIntent("operators.GetOperatorBySigningKey")
	dbRead := func() (any, time.Duration, error) {
		cacheDuration := time.Duration(time.Millisecond * 30000)
		row := q.GetConn().WQueryRow(qctx, "operators.GetOperatorBySigningKey", getOperatorBySigningKey, signingKey)
		var i *Operators = new(Operators)
		err := row.Scan(
			&i.Address,
			&i.SigningKey,
			&i.P2pKey,
			&i.RegisteredAtBlockNumber,
			&i.ActiveEpoch,
			&i.ExitEpoch,
			&i.IsActive,
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
	err := q.GetCache().GetWithTtl(qctx, "operators:GetOperatorBySigningKey:"+hashIfLong(fmt.Sprintf("%+v", signingKey)), &i, dbRead, false, false)
	if err != nil {
		return nil, err
	}

	return i, err
}

const isActiveOperator = `-- name: IsActiveOperator :one
SELECT EXISTS (
    SELECT 1 FROM operators
    WHERE LOWER(p2p_key) = LOWER($1)
    AND is_active = true
) as is_active
`

// -- cache: 168h
// -- timeout: 500ms
func (q *Queries) IsActiveOperator(ctx context.Context, lower string) (*bool, error) {
	return _IsActiveOperator(ctx, q.AsReadOnly(), lower)
}

func (q *ReadOnlyQueries) IsActiveOperator(ctx context.Context, lower string) (*bool, error) {
	return _IsActiveOperator(ctx, q, lower)
}

func _IsActiveOperator(ctx context.Context, q CacheQuerierConn, lower string) (*bool, error) {
	qctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	q.GetConn().CountIntent("operators.IsActiveOperator")
	dbRead := func() (any, time.Duration, error) {
		cacheDuration := time.Duration(time.Millisecond * 604800000)
		row := q.GetConn().WQueryRow(qctx, "operators.IsActiveOperator", isActiveOperator, lower)
		var is_active *bool = new(bool)
		err := row.Scan(is_active)
		if err == pgx.ErrNoRows {
			return (*bool)(nil), cacheDuration, nil
		}
		return is_active, cacheDuration, err
	}
	if q.GetCache() == nil {
		is_active, _, err := dbRead()
		return is_active.(*bool), err
	}

	var is_active *bool
	err := q.GetCache().GetWithTtl(qctx, "operators:IsActiveOperator:"+hashIfLong(fmt.Sprintf("%+v", lower)), &is_active, dbRead, false, false)
	if err != nil {
		return nil, err
	}

	return is_active, err
}

const listActiveOperators = `-- name: ListActiveOperators :many
SELECT address, signing_key, p2p_key, registered_at_block_number, active_epoch, exit_epoch, is_active, weight, created_at, updated_at FROM operators
WHERE is_active = true
`

// -- timeout: 1s
func (q *Queries) ListActiveOperators(ctx context.Context) ([]Operators, error) {
	return _ListActiveOperators(ctx, q.AsReadOnly())
}

func (q *ReadOnlyQueries) ListActiveOperators(ctx context.Context) ([]Operators, error) {
	return _ListActiveOperators(ctx, q)
}

func _ListActiveOperators(ctx context.Context, q CacheQuerierConn) ([]Operators, error) {
	qctx, cancel := context.WithTimeout(ctx, time.Millisecond*1000)
	defer cancel()
	q.GetConn().CountIntent("operators.ListActiveOperators")
	rows, err := q.GetConn().WQuery(qctx, "operators.ListActiveOperators", listActiveOperators)
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
			&i.P2pKey,
			&i.RegisteredAtBlockNumber,
			&i.ActiveEpoch,
			&i.ExitEpoch,
			&i.IsActive,
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

const listAllOperators = `-- name: ListAllOperators :many
SELECT address, signing_key, p2p_key, registered_at_block_number, active_epoch, exit_epoch, is_active, weight, created_at, updated_at FROM operators
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
			&i.P2pKey,
			&i.RegisteredAtBlockNumber,
			&i.ActiveEpoch,
			&i.ExitEpoch,
			&i.IsActive,
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
RETURNING address, signing_key, p2p_key, registered_at_block_number, active_epoch, exit_epoch, is_active, weight, created_at, updated_at
`

type UpdateOperatorExitEpochParams struct {
	Address   string `json:"address"`
	ExitEpoch uint32 `json:"exit_epoch"`
}

// -- invalidate: [GetOperatorByAddress, GetOperatorBySigningKey, GetOperatorByP2PKey, IsActiveOperator]
// -- timeout: 500ms
func (q *Queries) UpdateOperatorExitEpoch(ctx context.Context, arg UpdateOperatorExitEpochParams, getOperatorByAddress *string, getOperatorBySigningKey *string, getOperatorByP2PKey *string, isActiveOperator *string) (*Operators, error) {
	return _UpdateOperatorExitEpoch(ctx, q, arg, getOperatorByAddress, getOperatorBySigningKey, getOperatorByP2PKey, isActiveOperator)
}

func _UpdateOperatorExitEpoch(ctx context.Context, q CacheWGConn, arg UpdateOperatorExitEpochParams, getOperatorByAddress *string, getOperatorBySigningKey *string, getOperatorByP2PKey *string, isActiveOperator *string) (*Operators, error) {
	qctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	row := q.GetConn().WQueryRow(qctx, "operators.UpdateOperatorExitEpoch", updateOperatorExitEpoch, arg.Address, arg.ExitEpoch)
	var i *Operators = new(Operators)
	err := row.Scan(
		&i.Address,
		&i.SigningKey,
		&i.P2pKey,
		&i.RegisteredAtBlockNumber,
		&i.ActiveEpoch,
		&i.ExitEpoch,
		&i.IsActive,
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
		anyErr := make(chan error, 4)
		var wg sync.WaitGroup
		wg.Add(4)
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
		go func() {
			defer wg.Done()
			if getOperatorBySigningKey != nil {
				key := "operators:GetOperatorBySigningKey:" + hashIfLong(fmt.Sprintf("%+v", (*getOperatorBySigningKey)))
				err = q.GetCache().Invalidate(ctx, key)
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Msgf(
						"Failed to invalidate: %s", key)
					anyErr <- err
				}
			}
		}()
		go func() {
			defer wg.Done()
			if getOperatorByP2PKey != nil {
				key := "operators:GetOperatorByP2PKey:" + hashIfLong(fmt.Sprintf("%+v", (*getOperatorByP2PKey)))
				err = q.GetCache().Invalidate(ctx, key)
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Msgf(
						"Failed to invalidate: %s", key)
					anyErr <- err
				}
			}
		}()
		go func() {
			defer wg.Done()
			if isActiveOperator != nil {
				key := "operators:IsActiveOperator:" + hashIfLong(fmt.Sprintf("%+v", (*isActiveOperator)))
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
SET is_active = $2,
    weight = $3,
    signing_key = $4,
    p2p_key = $5,
    updated_at = NOW()
WHERE address = $1
RETURNING address, signing_key, p2p_key, registered_at_block_number, active_epoch, exit_epoch, is_active, weight, created_at, updated_at
`

type UpdateOperatorStateParams struct {
	Address    string         `json:"address"`
	IsActive   bool           `json:"is_active"`
	Weight     pgtype.Numeric `json:"weight"`
	SigningKey string         `json:"signing_key"`
	P2pKey     string         `json:"p2p_key"`
}

// -- invalidate: [GetOperatorByAddress, GetOperatorBySigningKey, GetOperatorByP2PKey, IsActiveOperator]
// -- timeout: 500ms
func (q *Queries) UpdateOperatorState(ctx context.Context, arg UpdateOperatorStateParams, getOperatorByAddress *string, getOperatorBySigningKey *string, getOperatorByP2PKey *string, isActiveOperator *string) (*Operators, error) {
	return _UpdateOperatorState(ctx, q, arg, getOperatorByAddress, getOperatorBySigningKey, getOperatorByP2PKey, isActiveOperator)
}

func _UpdateOperatorState(ctx context.Context, q CacheWGConn, arg UpdateOperatorStateParams, getOperatorByAddress *string, getOperatorBySigningKey *string, getOperatorByP2PKey *string, isActiveOperator *string) (*Operators, error) {
	qctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	row := q.GetConn().WQueryRow(qctx, "operators.UpdateOperatorState", updateOperatorState,
		arg.Address,
		arg.IsActive,
		arg.Weight,
		arg.SigningKey,
		arg.P2pKey)
	var i *Operators = new(Operators)
	err := row.Scan(
		&i.Address,
		&i.SigningKey,
		&i.P2pKey,
		&i.RegisteredAtBlockNumber,
		&i.ActiveEpoch,
		&i.ExitEpoch,
		&i.IsActive,
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
		anyErr := make(chan error, 4)
		var wg sync.WaitGroup
		wg.Add(4)
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
		go func() {
			defer wg.Done()
			if getOperatorBySigningKey != nil {
				key := "operators:GetOperatorBySigningKey:" + hashIfLong(fmt.Sprintf("%+v", (*getOperatorBySigningKey)))
				err = q.GetCache().Invalidate(ctx, key)
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Msgf(
						"Failed to invalidate: %s", key)
					anyErr <- err
				}
			}
		}()
		go func() {
			defer wg.Done()
			if getOperatorByP2PKey != nil {
				key := "operators:GetOperatorByP2PKey:" + hashIfLong(fmt.Sprintf("%+v", (*getOperatorByP2PKey)))
				err = q.GetCache().Invalidate(ctx, key)
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Msgf(
						"Failed to invalidate: %s", key)
					anyErr <- err
				}
			}
		}()
		go func() {
			defer wg.Done()
			if isActiveOperator != nil {
				key := "operators:IsActiveOperator:" + hashIfLong(fmt.Sprintf("%+v", (*isActiveOperator)))
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
    p2p_key,
    registered_at_block_number,
    active_epoch,
    exit_epoch,
    is_active,
    weight,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()
)
ON CONFLICT (address) DO UPDATE
SET 
    signing_key = EXCLUDED.signing_key,
    p2p_key = EXCLUDED.p2p_key,
    registered_at_block_number = EXCLUDED.registered_at_block_number,
    active_epoch = EXCLUDED.active_epoch,
    exit_epoch = EXCLUDED.exit_epoch,
    is_active = EXCLUDED.is_active,
    weight = EXCLUDED.weight,
    updated_at = NOW()
RETURNING address, signing_key, p2p_key, registered_at_block_number, active_epoch, exit_epoch, is_active, weight, created_at, updated_at
`

type UpsertOperatorParams struct {
	Address                 string         `json:"address"`
	SigningKey              string         `json:"signing_key"`
	P2pKey                  string         `json:"p2p_key"`
	RegisteredAtBlockNumber uint64         `json:"registered_at_block_number"`
	ActiveEpoch             uint32         `json:"active_epoch"`
	ExitEpoch               uint32         `json:"exit_epoch"`
	IsActive                bool           `json:"is_active"`
	Weight                  pgtype.Numeric `json:"weight"`
}

// -- invalidate: [GetOperatorByAddress, GetOperatorBySigningKey, GetOperatorByP2PKey, IsActiveOperator]
// -- timeout: 500ms
func (q *Queries) UpsertOperator(ctx context.Context, arg UpsertOperatorParams, getOperatorByAddress *string, getOperatorBySigningKey *string, getOperatorByP2PKey *string, isActiveOperator *string) (*Operators, error) {
	return _UpsertOperator(ctx, q, arg, getOperatorByAddress, getOperatorBySigningKey, getOperatorByP2PKey, isActiveOperator)
}

func _UpsertOperator(ctx context.Context, q CacheWGConn, arg UpsertOperatorParams, getOperatorByAddress *string, getOperatorBySigningKey *string, getOperatorByP2PKey *string, isActiveOperator *string) (*Operators, error) {
	qctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	row := q.GetConn().WQueryRow(qctx, "operators.UpsertOperator", upsertOperator,
		arg.Address,
		arg.SigningKey,
		arg.P2pKey,
		arg.RegisteredAtBlockNumber,
		arg.ActiveEpoch,
		arg.ExitEpoch,
		arg.IsActive,
		arg.Weight)
	var i *Operators = new(Operators)
	err := row.Scan(
		&i.Address,
		&i.SigningKey,
		&i.P2pKey,
		&i.RegisteredAtBlockNumber,
		&i.ActiveEpoch,
		&i.ExitEpoch,
		&i.IsActive,
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
		anyErr := make(chan error, 4)
		var wg sync.WaitGroup
		wg.Add(4)
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
		go func() {
			defer wg.Done()
			if getOperatorBySigningKey != nil {
				key := "operators:GetOperatorBySigningKey:" + hashIfLong(fmt.Sprintf("%+v", (*getOperatorBySigningKey)))
				err = q.GetCache().Invalidate(ctx, key)
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Msgf(
						"Failed to invalidate: %s", key)
					anyErr <- err
				}
			}
		}()
		go func() {
			defer wg.Done()
			if getOperatorByP2PKey != nil {
				key := "operators:GetOperatorByP2PKey:" + hashIfLong(fmt.Sprintf("%+v", (*getOperatorByP2PKey)))
				err = q.GetCache().Invalidate(ctx, key)
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Msgf(
						"Failed to invalidate: %s", key)
					anyErr <- err
				}
			}
		}()
		go func() {
			defer wg.Done()
			if isActiveOperator != nil {
				key := "operators:IsActiveOperator:" + hashIfLong(fmt.Sprintf("%+v", (*isActiveOperator)))
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

//// auto generated functions

func (q *Queries) Dump(ctx context.Context, beforeDump ...BeforeDump) ([]byte, error) {
	sql := "SELECT address,signing_key,p2p_key,registered_at_block_number,active_epoch,exit_epoch,is_active,weight,created_at,updated_at FROM \"operators\" ORDER BY address,signing_key,p2p_key,is_active,created_at,updated_at ASC;"
	rows, err := q.db.WQuery(ctx, "operators.Dump", sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Operators
	for rows.Next() {
		var v Operators
		if err := rows.Scan(&v.Address, &v.SigningKey, &v.P2pKey, &v.RegisteredAtBlockNumber, &v.ActiveEpoch, &v.ExitEpoch, &v.IsActive, &v.Weight, &v.CreatedAt, &v.UpdatedAt); err != nil {
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
	sql := "INSERT INTO \"operators\" (address,signing_key,p2p_key,registered_at_block_number,active_epoch,exit_epoch,is_active,weight,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10);"
	rows := make([]Operators, 0)
	err := json.Unmarshal(data, &rows)
	if err != nil {
		return err
	}
	for _, row := range rows {
		_, err := q.db.WExec(ctx, "operators.Load", sql, row.Address, row.SigningKey, row.P2pKey, row.RegisteredAtBlockNumber, row.ActiveEpoch, row.ExitEpoch, row.IsActive, row.Weight, row.CreatedAt, row.UpdatedAt)
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
