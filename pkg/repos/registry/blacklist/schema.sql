CREATE TABLE IF NOT EXISTS blacklist (
    id              BIGSERIAL       PRIMARY KEY,

    peer_id         VARCHAR(255)    NOT NULL,    -- libp2p peer ID
    ip              VARCHAR(45)     NOT NULL,    -- 节点的公网IP
    
    reason          TEXT           NULL,         -- 封禁原因
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ    NULL CHECK (expires_at > created_at),

    CONSTRAINT valid_peer_id CHECK (
        peer_id ~ '^12D3KooW[1-9A-HJ-NP-Za-km-z]{44,48}$'  -- libp2p peer ID
    ),
    CONSTRAINT valid_ip CHECK (
        ip ~ '^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$' OR          -- IPv4
        ip ~ '^[0-9a-fA-F:]+$'                             -- IPv6
    ),

    CONSTRAINT unique_peer_ip UNIQUE (peer_id, ip)
);

-- index
CREATE INDEX IF NOT EXISTS blacklist_peer_id_idx ON blacklist(peer_id);
CREATE INDEX IF NOT EXISTS blacklist_ip_idx ON blacklist(ip);
CREATE INDEX IF NOT EXISTS blacklist_created_at_idx ON blacklist(created_at);
CREATE INDEX IF NOT EXISTS blacklist_expires_at_idx ON blacklist(expires_at);

