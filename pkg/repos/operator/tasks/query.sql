-- name: CreateTask :one
-- -- timeout: 500ms

INSERT INTO tasks (
    task_id,
    target_address,
    chain_id,
    block_number,
    timestamp,
    epoch,
    key,
    value,
    status,
    required_confirmations
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetTaskByID :one
-- -- timeout: 500ms
SELECT * FROM tasks
WHERE task_id = $1;

-- name: UpdateTaskStatus :one
-- -- timeout: 500ms

UPDATE tasks
SET status = $2,
    updated_at = NOW()
WHERE task_id = $1
RETURNING *;

-- name: ListPendingTasks :many
-- -- timeout: 500ms
SELECT * FROM tasks
WHERE status = 'pending'
ORDER BY created_at ASC;

-- name: CleanupOldTasks :exec
-- -- timeout: 5s
DELETE FROM tasks
WHERE created_at < NOW() - INTERVAL '24 hours'
AND status IN ('completed');

-- name: ListConfirmingTasks :many
-- -- timeout: 500ms
SELECT * FROM tasks 
WHERE status = 'confirming'
ORDER BY created_at DESC;

-- name: UpdateTaskCompleted :exec
-- -- timeout: 500ms


UPDATE tasks
SET status = 'completed',
    updated_at = NOW()
WHERE task_id = $1;

-- name: UpdateTaskToPending :exec
-- -- timeout: 500ms


UPDATE tasks 
SET status = 'pending', 
    updated_at = NOW()
WHERE task_id = $1;

-- name: ListAllTasks :many
-- -- timeout: 500ms
SELECT * FROM tasks ORDER BY created_at DESC;

-- name: IncrementRetryCount :one
-- -- timeout: 500ms


UPDATE tasks
SET retry_count = retry_count + 1,
    updated_at = NOW()
WHERE task_id = $1
RETURNING *;

-- name: DeleteTaskByID :exec
-- -- timeout: 500ms

DELETE FROM tasks
WHERE task_id = $1;

