-- name: IncrementViolationCount :one
-- -- timeout: 500ms
-- -- invalidate: IsBlocked
INSERT INTO blacklist (
    peer_id,
    violation_count,
    expires_at
) VALUES (
    @peer_id,
    @violation_count,
    @expires_at
)
ON CONFLICT (peer_id) DO UPDATE
SET 
    violation_count = blacklist.violation_count + EXCLUDED.violation_count,
    expires_at = EXCLUDED.expires_at,
    updated_at = NOW()
RETURNING *;

-- name: UnblockNode :exec
-- -- timeout: 500ms
-- -- invalidate: IsBlocked
DELETE FROM blacklist 
WHERE peer_id = @peer_id;

-- name: IsBlocked :one
-- -- cache: 24h
-- -- timeout: 100ms
SELECT EXISTS (
    SELECT 1 FROM blacklist
    WHERE peer_id = @peer_id
    AND violation_count >= 3
    AND (expires_at IS NULL OR expires_at > NOW())
) as is_blocked;

-- name: ListBlacklist :many
-- -- timeout: 500ms
SELECT * FROM blacklist
WHERE expires_at IS NULL OR expires_at > NOW()
ORDER BY created_at DESC
LIMIT COALESCE(sqlc.arg('limit')::int, 100)
OFFSET COALESCE(sqlc.arg('offset')::int, 0);

-- name: CleanExpiredBlocks :exec
-- -- timeout: 5s
-- -- invalidate: IsBlocked
DELETE FROM blacklist
WHERE expires_at < NOW();


