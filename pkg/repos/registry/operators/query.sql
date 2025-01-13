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