-- name: BlockNode :one
-- -- timeout: 500ms
-- -- invalidate: IsBlocked
INSERT INTO blacklist (
    peer_id,
    ip,
    reason,
    expires_at
) VALUES (
    @peer_id,
    @ip,
    @reason,
    @expires_at
)
ON CONFLICT (peer_id, ip) DO UPDATE
SET 
    reason = EXCLUDED.reason,
    expires_at = EXCLUDED.expires_at
RETURNING *;

-- name: UnblockNode :exec
-- -- timeout: 500ms
-- -- invalidate: IsBlocked
DELETE FROM blacklist 
WHERE peer_id = @peer_id AND ip = @ip;

-- name: IsBlocked :one
-- -- timeout: 100ms
-- -- cache: 24h
SELECT EXISTS (
    SELECT 1 FROM blacklist
    WHERE (peer_id = @peer_id OR ip = @ip)
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
DELETE FROM blacklist
WHERE expires_at < NOW();

-- name: RefreshIDSerial :exec
-- -- timeout: 300ms
SELECT setval(pg_get_serial_sequence('blacklist', 'id'), (SELECT MAX(id) FROM blacklist)+1, false);

