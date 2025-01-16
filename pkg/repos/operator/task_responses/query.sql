-- name: CreateTaskResponse :one
INSERT INTO task_responses (
    task_id,
    operator_address,
    signature,
    epoch,
    chain_id,
    target_address,
    key,
    value,
    block_number,
    timestamp,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetTaskResponses :many
-- Get all responses for a specific task
SELECT * FROM task_responses
WHERE task_id = $1;

-- name: GetTaskResponsesByStatus :many
-- Get all responses for a task with specific status
SELECT * FROM task_responses
WHERE task_id = $1 AND status = $2;

-- name: GetOperatorResponse :one
-- Get specific operator's response for a task
SELECT * FROM task_responses
WHERE task_id = $1 AND operator_address = $2;

-- name: UpdateResponseStatus :one
-- Update response status
UPDATE task_responses
SET status = $3
WHERE task_id = $1 AND operator_address = $2
RETURNING *;

-- name: GetResponseCount :one
-- Get count of responses for a task
SELECT COUNT(*) FROM task_responses
WHERE task_id = $1 AND status = $2; 