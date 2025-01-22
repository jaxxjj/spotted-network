-- name: GetOperatorByAddress :one
-- Get operator information by address
SELECT * FROM operators
WHERE address = $1;

-- name: CreateOperator :one
-- Create a new operator record with waitingJoin status
INSERT INTO operators (
    address,
    signing_key,
    registered_at_block_number,
    registered_at_timestamp,
    active_epoch,
    status,
    weight
) VALUES ($1, $2, $3, $4, $5, 'waitingJoin', $6)
RETURNING *;

-- name: UpdateOperatorStatus :one
-- Update operator status
UPDATE operators
SET status = $2,
    updated_at = NOW()
WHERE address = $1
RETURNING *;

-- name: VerifyOperatorStatus :one
-- Verify operator status and signing key for join request
SELECT status, signing_key 
FROM operators
WHERE address = $1 AND status = 'waitingJoin';

-- name: ListOperatorsByStatus :many
-- Get all operators with a specific status
SELECT * FROM operators
WHERE status = $1
ORDER BY created_at DESC;

-- name: ListAllOperators :many
-- Get all operators regardless of status
SELECT * FROM operators
ORDER BY created_at DESC;

-- name: UpdateOperatorState :one
-- Update operator status and weight
UPDATE operators
SET status = $2,
    weight = $3,
    updated_at = NOW()
WHERE address = $1
RETURNING *;

-- name: UpdateOperatorExitEpoch :one
-- Update operator exit epoch and status
UPDATE operators
SET exit_epoch = $2,
    status = $3,
    updated_at = NOW()
WHERE address = $1
RETURNING *;

-- name: UpsertOperator :one
-- Insert or update operator record
INSERT INTO operators (
    address,
    signing_key,
    registered_at_block_number,
    registered_at_timestamp,
    active_epoch,
    status,
    weight,
    exit_epoch
) VALUES ($1, $2, $3, $4, $5, 'waitingJoin', $6, $7)
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