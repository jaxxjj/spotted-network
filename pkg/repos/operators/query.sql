-- name: GetOperatorByAddress :one
-- -- cache: 10s
-- -- timeout: 500ms
SELECT * FROM operators
WHERE address = $1;

-- name: GetOperatorBySigningKey :one
-- -- cache: 10s
-- -- timeout: 500ms
SELECT * FROM operators
WHERE signing_key = $1;

-- name: GetOperatorByP2PKey :one
-- -- cache: 10s
-- -- timeout: 500ms
SELECT * FROM operators
WHERE LOWER(p2p_key) = LOWER($1);

-- name: IsActiveOperator :one
-- -- cache: 10s
-- -- timeout: 500ms
SELECT EXISTS (
    SELECT 1 FROM operators
    WHERE LOWER(p2p_key) = LOWER($1)
    AND is_active = true
) as is_active;

-- name: ListActiveOperators :many
-- -- timeout: 1s
SELECT * FROM operators
WHERE is_active = true;

-- name: ListAllOperators :many
-- -- timeout: 1s
SELECT * FROM operators;

-- name: UpdateOperatorState :one
-- -- invalidate: [GetOperatorByAddress, GetOperatorBySigningKey, GetOperatorByP2PKey, IsActiveOperator]
-- -- timeout: 500ms
UPDATE operators
SET is_active = $2,
    weight = $3,
    signing_key = $4,
    p2p_key = $5,
    updated_at = NOW()
WHERE address = $1
RETURNING *;

-- name: UpdateOperatorExitEpoch :one
-- -- invalidate: [GetOperatorByAddress, GetOperatorBySigningKey, GetOperatorByP2PKey, IsActiveOperator]
-- -- timeout: 500ms
UPDATE operators
SET exit_epoch = $2,
    updated_at = NOW()
WHERE address = $1
RETURNING *;

-- name: UpsertOperator :one
-- -- invalidate: [GetOperatorByAddress, GetOperatorBySigningKey, GetOperatorByP2PKey, IsActiveOperator]
-- -- timeout: 500ms
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
RETURNING *;


