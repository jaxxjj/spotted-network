-- name: CreateConsensusResponse :one
INSERT INTO consensus_responses (
    task_id,
    epoch,
    value,
    block_number,
    chain_id,
    target_address,
    key,
    aggregated_signatures,
    operator_signatures,
    total_weight,
    consensus_reached_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetConsensusResponseByTaskId :one
SELECT * FROM consensus_responses
WHERE task_id = $1 LIMIT 1;

-- name: GetConsensusResponseByRequest :one
SELECT * FROM consensus_responses
WHERE target_address = $1 
AND chain_id = $2 
AND block_number = $3 
AND key = $4 
LIMIT 1;

-- name: UpdateConsensusResponse :one
UPDATE consensus_responses
SET consensus_reached_at = $2,
    aggregated_signatures = $3,
    operator_signatures = $4
WHERE task_id = $1
RETURNING *;

-- name: ListPendingConsensus :many
SELECT * FROM consensus_responses
WHERE status = 'pending'
ORDER BY created_at DESC;

-- name: DeleteConsensusResponse :exec
DELETE FROM consensus_responses
WHERE task_id = $1; 