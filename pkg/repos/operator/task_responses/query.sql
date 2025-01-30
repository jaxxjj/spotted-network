-- name: CreateTaskResponse :one
-- -- timeout: 500ms
INSERT INTO task_responses (
    task_id,
    operator_address,
    signing_key,
    signature,
    epoch,
    chain_id,
    target_address,
    key,
    value,
    block_number,
    timestamp
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetTaskResponse :one
-- -- timeout: 500ms
SELECT * FROM task_responses
WHERE task_id = $1 AND operator_address = $2;

-- name: ListTaskResponses :many
-- -- timeout: 1s
SELECT * FROM task_responses
WHERE task_id = $1;

-- name: ListOperatorResponses :many
-- -- timeout: 1s
SELECT * FROM task_responses
WHERE operator_address = $1
ORDER BY submitted_at DESC
LIMIT $2;

-- name: DeleteTaskResponse :exec
-- -- timeout: 500ms
DELETE FROM task_responses
WHERE task_id = $1 AND operator_address = $2; 