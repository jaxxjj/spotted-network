-- name: GetOperatorByAddress :one
-- Get operator information by address
-- -- cache: 168h
-- -- timeout: 500ms
SELECT * FROM operators
WHERE address = $1;

-- name: CreateOperator :one
-- Create a new operator record with inactive status
-- -- invalidate: GetOperatorByAddress
-- -- timeout: 500ms
INSERT INTO operators (
    address,
    signing_key,
    registered_at_block_number,
    registered_at_timestamp,
    active_epoch,
    status,
    weight
) VALUES ($1, $2, $3, $4, $5, 'inactive', $6)
RETURNING *;

-- name: UpdateOperatorStatus :one
-- -- invalidate: GetOperatorByAddress
-- -- timeout: 500ms
UPDATE operators
SET status = $2,
    updated_at = NOW()
WHERE address = $1
RETURNING *;

-- name: VerifyOperatorStatus :one
-- -- timeout: 500ms
SELECT status, signing_key 
FROM operators
WHERE address = $1 AND status = 'active';

-- name: ListOperatorsByStatus :many
-- -- timeout: 1s
SELECT * FROM operators
WHERE status = $1
ORDER BY created_at DESC;

-- name: ListAllOperators :many
-- -- timeout: 1s
SELECT * FROM operators
ORDER BY created_at DESC;

-- name: UpdateOperatorState :one
-- -- invalidate: GetOperatorByAddress
-- -- timeout: 500ms
UPDATE operators
SET status = $2,
    weight = $3,
    signing_key = $4,
    updated_at = NOW()
WHERE address = $1
RETURNING *;

-- name: UpdateOperatorExitEpoch :one
-- -- invalidate: GetOperatorByAddress
-- -- timeout: 500ms
UPDATE operators
SET exit_epoch = $2,
    updated_at = NOW()
WHERE address = $1
RETURNING *;

-- name: UpsertOperator :one
-- -- invalidate: GetOperatorByAddress
-- -- timeout: 500ms
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
RETURNING *; 

