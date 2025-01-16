-- name: CreateConsensusResponse :one
INSERT INTO consensus_responses (
    task_id,
    epoch,
    status,
    aggregated_signatures,
    operator_signatures,
    consensus_reached_at
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetConsensusResponse :one
SELECT * FROM consensus_responses
WHERE task_id = $1;

-- name: UpdateConsensusStatus :one
UPDATE consensus_responses
SET status = $2,
    consensus_reached_at = $3,
    aggregated_signatures = $4,
    operator_signatures = $5
WHERE task_id = $1
RETURNING *;

-- name: ListPendingConsensus :many
SELECT * FROM consensus_responses
WHERE status = 'pending'
ORDER BY created_at DESC;

-- name: DeleteConsensusResponse :exec
DELETE FROM consensus_responses
WHERE task_id = $1; 