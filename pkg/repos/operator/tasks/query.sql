-- name: CreateTask :one
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
SELECT * FROM tasks
WHERE task_id = $1;

-- name: UpdateTaskStatus :one
UPDATE tasks
SET status = $2,
    updated_at = NOW()
WHERE task_id = $1
RETURNING *;

-- name: UpdateTaskValue :one
UPDATE tasks
SET value = $2,
    status = 'pending',
    updated_at = NOW()
WHERE task_id = $1
RETURNING *;

-- name: ListPendingTasks :many
SELECT * FROM tasks
WHERE status = 'pending'
ORDER BY created_at ASC;

-- name: CleanupOldTasks :exec
DELETE FROM tasks
WHERE created_at < NOW() - INTERVAL '24 hours'
AND status IN ('completed');

-- name: ListConfirmingTasks :many
SELECT * FROM tasks 
WHERE status = 'confirming'
ORDER BY created_at DESC;

-- name: UpdateTaskCompleted :exec
UPDATE tasks
SET status = 'completed',
    updated_at = NOW()
WHERE task_id = $1;

-- name: UpdateTaskToPending :exec
UPDATE tasks 
SET status = 'pending', 
    updated_at = NOW()
WHERE task_id = $1;

-- name: ListAllTasks :many
SELECT * FROM tasks ORDER BY created_at DESC;

-- name: IncrementRetryCount :one
UPDATE tasks
SET retry_count = retry_count + 1,
    updated_at = NOW()
WHERE task_id = $1
RETURNING *;

-- name: DeleteTasksByRetryCount :exec
DELETE FROM tasks
WHERE retry_count >= $1
AND status = 'pending';