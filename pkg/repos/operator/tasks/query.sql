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
    expire_time,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetTaskByID :one
SELECT * FROM tasks
WHERE task_id = $1;

-- name: UpdateTaskStatus :one
UPDATE tasks
SET status = $2,
    retries = CASE 
        WHEN $2 = 'pending' THEN retries + 1
        ELSE retries
    END
WHERE task_id = $1
RETURNING *;

-- name: UpdateTaskValue :one
UPDATE tasks
SET value = $2,
    status = 'pending'
WHERE task_id = $1
RETURNING *;

-- name: ListPendingTasks :many
SELECT * FROM tasks
WHERE status = 'pending'
AND expire_time > NOW()
ORDER BY created_at ASC;

-- name: ListExpiredTasks :many
SELECT * FROM tasks
WHERE status = 'pending'
AND expire_time <= NOW()
ORDER BY created_at ASC;

-- name: CleanupOldTasks :exec
DELETE FROM tasks
WHERE created_at < NOW() - INTERVAL '24 hours'
AND status IN ('completed', 'expired', 'failed'); 