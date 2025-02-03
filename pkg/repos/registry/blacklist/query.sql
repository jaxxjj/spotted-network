-- name: UpsertBlacklist :one
-- Add or update a blacklist entry
INSERT INTO blacklists (
    type,
    value,
    reason,
    created_by,
    expires_at
) VALUES (
    @type::blacklist_type,
    @value,
    @reason,
    @created_by,
    @expires_at
)
ON CONFLICT (type, value) WHERE deleted_at IS NULL DO UPDATE
SET 
    reason = EXCLUDED.reason,
    created_by = EXCLUDED.created_by,
    expires_at = EXCLUDED.expires_at,
    created_at = NOW()
RETURNING *;

-- name: BulkUpsertBlacklist :copyfrom
-- Bulk add or update blacklist entries
INSERT INTO blacklists (
    type,
    value,
    reason,
    created_by,
    expires_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
);

-- name: RemoveFromBlacklist :exec
-- Soft delete a blacklist entry
UPDATE blacklists 
SET deleted_at = NOW()
WHERE type = @type::blacklist_type 
  AND value = @value
  AND deleted_at IS NULL;

-- name: IsBlacklisted :one
-- Check if an entry is blacklisted
-- -- timeout: 100ms
-- -- cache: 1m
SELECT EXISTS (
    SELECT 1 FROM active_blacklists
    WHERE type = @type::blacklist_type 
      AND value = @value
) as is_blacklisted;

-- name: ListActiveBlacklists :many
-- Get all active blacklist entries
-- -- timeout: 500ms
-- -- cache: 1m
SELECT * FROM active_blacklists
WHERE type = COALESCE(sqlc.narg('type')::blacklist_type, type)
ORDER BY created_at DESC;

-- name: GetBlacklistByValue :one
-- Get a specific blacklist entry
-- -- timeout: 100ms
-- -- cache: 1m
SELECT * FROM active_blacklists
WHERE type = @type::blacklist_type 
  AND value = @value;

-- name: CleanExpiredBlacklists :exec
-- Remove expired entries
DELETE FROM blacklists
WHERE expires_at < NOW()
  AND deleted_at IS NULL; 