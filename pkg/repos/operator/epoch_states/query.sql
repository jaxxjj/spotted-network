-- name: UpsertEpochState :one
-- -- invalidate: GetEpochState
-- -- invalidate: GetLatestEpochState
-- -- invalidate: ListEpochStates
INSERT INTO epoch_states (
    epoch_number,
    block_number,
    minimum_weight,
    total_weight,
    threshold_weight,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
ON CONFLICT (epoch_number) DO UPDATE
SET
    block_number = EXCLUDED.block_number,
    minimum_weight = EXCLUDED.minimum_weight,
    total_weight = EXCLUDED.total_weight,
    threshold_weight = EXCLUDED.threshold_weight,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: GetEpochState :one
-- -- cache: 7d
SELECT * FROM epoch_states
WHERE epoch_number = $1;

-- name: GetLatestEpochState :one
-- -- cache: 1h
SELECT * FROM epoch_states
ORDER BY epoch_number DESC
LIMIT 1;

-- name: ListEpochStates :many
-- -- cache: 7d
SELECT * FROM epoch_states
ORDER BY epoch_number DESC
LIMIT $1; 