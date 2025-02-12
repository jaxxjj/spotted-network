CREATE TABLE IF NOT EXISTS blacklist (
    peer_id VARCHAR(128) PRIMARY KEY,
    violation_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NULL CHECK (expires_at > created_at),

    CONSTRAINT valid_violation_count CHECK (violation_count >= 0)
);

CREATE INDEX IF NOT EXISTS idx_blacklist_violations ON blacklist(peer_id, violation_count, expires_at);

