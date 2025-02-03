-- Blacklist types enum
CREATE TYPE blacklist_type AS ENUM ('peer', 'ip', 'subnet');

-- Blacklist table for storing blocked entities
CREATE TABLE IF NOT EXISTS blacklists (
    id              BIGSERIAL       PRIMARY KEY,
    type            blacklist_type  NOT NULL,
    value           TEXT           NOT NULL, -- peer_id/ip/subnet
    reason          TEXT           NULL,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    created_by      TEXT           NULL,    -- who added this
    expires_at      TIMESTAMPTZ    NULL,    -- optional expiration
    deleted_at      TIMESTAMPTZ    NULL,    -- soft delete
    
    -- Ensure no duplicates for same type and value
    CONSTRAINT blacklists_type_value_unique UNIQUE NULLS NOT DISTINCT (type, value, deleted_at)
);

-- Indexes
CREATE INDEX IF NOT EXISTS blacklists_type_idx ON blacklists(type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS blacklists_created_at_idx ON blacklists(created_at);
CREATE INDEX IF NOT EXISTS blacklists_expires_at_idx ON blacklists(expires_at) WHERE deleted_at IS NULL;

-- View for active blacklists
CREATE OR REPLACE VIEW active_blacklists AS
SELECT * FROM blacklists 
WHERE deleted_at IS NULL 
  AND (expires_at IS NULL OR expires_at > NOW()); 